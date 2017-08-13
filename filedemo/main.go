package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bradleypeabody/gocaveman"
	"github.com/bradleypeabody/gocaveman/editor"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var PROGRAM_NAME = "filedemo"

func main() {

	// flags/viper (kingpin is probably overkill, just providing a -k CMD option is probably all the subcmd support we need)
	// start building out handler chain here
	// make a views, includes, etc. folder
	// start server and render
	// incprorate menus
	// make menu editor
	// service support...
	// work it

	pflag.StringP("http-listen", "l", ":8080", "IP:Port to listen on for HTTP")
	pflag.StringP("https-listen", "s", ":8443", "IP:Port to listen on for HTTPS")
	pflag.String("https-cert", "", "Path to (or contents of) HTTPS certificate in PEM format (not influenced by root-dir)")
	pflag.String("https-key", "", "Path to (or contents of) HTTPS key in PEM format (not influenced by root-dir)")
	pflag.StringP("root-dir", "r", ".", "Root directory for relative paths")
	pflag.String("static-path", "static", "Directory for static files (resolved within root-dir)")
	pflag.String("views-path", "views", "Directory for view template files (resolved within root-dir)")
	pflag.String("includes-path", "includes", "Directory for included template files (resolved within root-dir)")
	pflag.StringP("log-file", "f", "", "Path to log file, stdout/stderr is used if not specified (not influenced by root-dir)")
	cmd := pflag.StringP("cmd", "k", "", "Run service command instead of normal execution")

	if runtime.GOOS == "windows" {

	} else {
		pflag.String("pid-file", "/var/run/"+PROGRAM_NAME+".pid", "Write out process id file to this path during execution")
		pflag.String("uid", "", "Become this user during execution")
		pflag.String("gid", "", "Become this group during execution")
		pflag.Bool("syslog", false, "Log to syslog")
	}

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// viper config file scanning stuff

	switch *cmd {
	case "start":
		// options: uid, gid, pidfile
		// we also need the dup2 hack for capturing panic's to log file (see if panic capture is better done with a simple defer and log statement)
	case "stop":
	case "restart":
	case "status":
	case "install":
	case "uninstall":
	case "":
	default:
		log.Fatalf("Invalid command %q", *cmd)
	}

	// logging setup
	var logger lumberjack.Logger
	logFile := viper.GetString("log-file")
	if logFile != "" {
		logger.Filename = logFile
	}
	if viper.Get("logger") != nil {
		// FIXME: this is going to need mapstructure struct tags on lumberjack.Logger in order to work correctly with snake_case - contribute this fix to lumberjack
		err := viper.UnmarshalKey("logger", &logger)
		if err != nil {
			log.Fatal(err)
		}
	}
	if logger.Filename != "" {
		log.SetOutput(&logger)
	}
	if viper.GetBool("syslog") {
		err := gocaveman.LogToSyslog(PROGRAM_NAME)
		if err != nil {
			log.Fatal(err)
		}
	}

	// figure out root
	rootDir := viper.GetString("root-dir")
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using root-dir %q", rootDir)
	rootFs := afero.NewBasePathFs(afero.NewOsFs(), rootDir)
	if ok, err := afero.IsDir(rootFs, "/"); !ok {
		log.Fatalf("Invalid root directory %q: %v", rootDir, err)
	}

	// build handlers and whatever they need
	h, err := buildHandler(rootFs)
	if err != nil {
		log.Fatal(err)
	}

	// set up the http server
	httpServer := http.Server{
		Addr:           viper.GetString("http-listen"),
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// check for https cert+key pair
	tlsConfig := &tls.Config{}
	httpsCert := viper.GetString("http-cert")
	httpsKey := viper.GetString("http-key")
	if httpsCert != "" && httpsKey != "" {
		cert, err := gocaveman.LoadKeyPairAuto(httpsCert, httpsKey)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	// do https if at least one cert+key provided
	if len(tlsConfig.Certificates) > 0 {

		// copy the http server and change the https things
		httpsServer := httpServer
		httpsServer.Addr = ""
		httpsServer.Handler = gocaveman.CtxWithHTTPSHandler(h) // set "https"=true in the context

		l, err := net.Listen("tcp", viper.GetString("https-listen"))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Printf("Listening for HTTPS %q", l.Addr())
		go log.Fatal(httpsServer.Serve(l))

	} else {
		log.Printf("No SSL/TLS certificate+key pairs found, HTTPS disabled")
	}

	// always listen for HTTP
	log.Printf("Listening for HTTP %q", httpServer.Addr)
	log.Fatal(httpServer.ListenAndServe())

}

func buildHandler(rootFs afero.Fs) (http.Handler, error) {

	menus := gocaveman.NewJSONFileMenus(rootFs, "data/menus")

	editorFs := editor.NewDefaultPagesFs()

	// now build the actual handlers

	// gzip support
	gzipHandler := gocaveman.NewGzipHandler()

	// static stuff we need assigned to each context
	ctxMapHandler := gocaveman.NewCtxMapHandler(map[string]interface{}{
		"menus": menus,
	})

	// redirects
	defaultRedirectHandler := gocaveman.NewDefaultRedirects()

	// templates
	viewsFs := gocaveman.NewHTTPStackedFileSystem(
		afero.NewHttpFs(afero.NewBasePathFs(rootFs, viper.GetString("views-path"))),
		editorFs,
	)
	rendererHandler := gocaveman.NewDefaultRenderer(
		viewsFs,
		afero.NewHttpFs(afero.NewBasePathFs(rootFs, viper.GetString("includes-path"))),
	)

	// static files
	staticHandler := gocaveman.NewStaticFileServer(afero.NewBasePathFs(rootFs, viper.GetString("static-path")), "/")

	// assemble our handlers with the appropriate sequence/hierarchy
	h := gzipHandler.SetNextHandler(
		ctxMapHandler.SetNextHandler(
			gocaveman.NewHandlerChain(
				defaultRedirectHandler,
				rendererHandler,
				staticHandler,
				http.NotFoundHandler(),
			)))

	return h, nil

}
