package gocaveman

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mathrand "math/rand"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// TODO: we should add a good HTTPError here, it's pretty darned useful
type DefaultCacheHeadersHandler struct{}

func (h *DefaultCacheHeadersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: set based on if it looks like a page vs a resource and if it's combined or has cache id
	// hm, where is cacheid?  may need to go in this file too
}

// func DefaultCacheHeaders(w http.ResponseWriter, r *http.Request) {
// }

// TODO: gzipper
// TODO: dumper

// // StackableHandler allows you to add another handler
// type StackableHandler interface {
// 	http.Handler
// 	AddHandler(h http.Handler) StackableHandler
// }

// func NewStackableHandler(hs ...http.Handler) StackableHandler {

// for _, h := range ch {
// 	if r.Context().Err() != nil {
// 		return
// 	}
// 	h.ServeHTTP(w, r)
// }

// }

// type ChainedHandler interface {
// 	http.Handler
// 	SetNextHandler(http.Handler) ChainedHandler
// }

type ChainedHandler interface {
	http.Handler
	SetNextHandler(next http.Handler) (self ChainedHandler)
}

// // HandlerParent is meant to be embedded in handlers that need to modify
// // the request context and then let other handlers run with that modified context.
// type HandlerParent struct {
// 	ChildHandlers HandlerChain
// }

// func (hp *HandlerParent) AddHandler(h http.Handler) StackableHandler {
// 	hp.ChildHandlers = append(hp.ChildHandlers, h)
// }

// HandlerChain is a slice of http.Handler instances which is called in sequence
// until the request context has an error (which includes being cancelled because
// something was written to the output, the common case).
type HandlerChain []http.Handler

// func (ch HandlerChain) AddHandler(h http.Handler) StackableHandler {
// 	ret := append(ch, h)
// 	return ret
// }

func NewHandlerChain(hs ...http.Handler) HandlerChain {
	return HandlerChain(hs)
}

func (ch HandlerChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, h := range ch {
		if r.Context().Err() != nil {
			return
		}
		h.ServeHTTP(w, r)
	}
}

// type ChainWR struct {
// 	W http.ResponseWriter
// 	R *http.Request
// }

// // CtxChainWR returns the existing or new *WR for this context.
// func CtxChainWR(ctx context.Context) (*WR, context.Context) {
// 	wr, ok := ctx.Value("__wr").(*WR)
// 	if ok {
// 		return wr, ctx
// 	}
// 	wr = &WR{}
// 	ctx = context.WithValue(ctx, "__wr", wr)
// 	return wr, ctx
// }

// // CtxUpdateChainWR updates the chained response writer and request associated with this context.
// func CtxUpdateChainWR(ctx context.Context, w http.ResponseWriter, r *http.Request) {
// 	wr, _ := CtxChainWR(ctx)
// 	wr.W = w
// 	wr.R = r
// }

// func UpdateChainWR(w http.ResponseWriter, r *http.Request) {
// 	CtxUpdateChainWR(r.Context(), w, r)
// }

// FIXME: this should be http.FileSystem - we don't need write functionality for static file serving...
func NewStaticFileServer(fs afero.Fs, dir string) *StaticFileServer {
	return &StaticFileServer{Fs: fs, Dir: dir}
}

// StaticFileServer serves static files from a directory.
type StaticFileServer struct {
	Fs  afero.Fs
	Dir string
}

func (sfs *StaticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	p := path.Clean("/" + r.URL.Path)

	// no match on template, try for a static resource
	p1 := path.Join(sfs.Dir, p)
	f, err := sfs.Fs.Open(p1)
	if err != nil {

		// ignore not exist error, otherwise it's an error
		if !os.IsNotExist(err) {
			log.Printf("web file access error (p=%q): %v", p, err)
			http.Error(w, "web file access error", 500)
			return
		}

		// not exist here means not found
		return

	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		log.Printf("web file stat error (p=%q): %v", p, err)
		http.Error(w, "web file stat error", 500)
		return
	}

	// can't serve directory like this
	if st.IsDir() {
		return
	}

	w.Header().Set("content-type", GuessMIMEType(p))
	http.ServeContent(w, r, p, st.ModTime(), f)

}

// GuessMIMEType is a thin wrapper around mime.TypeByExtension(),
// but with some common sense defaults that are sometimes different/wrong
// on different platforms for no good reason.
func GuessMIMEType(p string) string {

	pext := path.Ext(p)

	switch pext {
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".html":
		return "text/html"
	}

	return mime.TypeByExtension(pext)

}

type ResponseWriterCloser interface {
	http.ResponseWriter
	io.Closer
}

// DebugResponseWriterCloser
type DebugResponseWriterCloser struct {
	http.ResponseWriter
	origw      http.ResponseWriter
	cancelFunc context.CancelFunc
}

func (w *DebugResponseWriterCloser) Write(p []byte) (int, error) {
	w.cancelFunc()
	return w.ResponseWriter.Write(p)
}

func (w *DebugResponseWriterCloser) WriteHeader(c int) {
	w.cancelFunc()
	w.ResponseWriter.WriteHeader(c)
}

// Close dumps the response using log.Printf
func (w *DebugResponseWriterCloser) Close() error {
	w.cancelFunc()

	rec := w.ResponseWriter.(*httptest.ResponseRecorder)

	res := rec.Result()
	b, _ := httputil.DumpResponse(res, true)
	log.Printf("---------- RESPONSE ----------\n%s", b)

	// replace headers
	h := w.origw.Header()
	for k := range h {
		h.Del(k)
	}
	newh := res.Header
	for k := range newh {
		h[k] = newh[k]
	}

	w.origw.WriteHeader(res.StatusCode)

	bodyb, _ := ioutil.ReadAll(res.Body)
	w.origw.Write(bodyb)

	return nil
}

func (w *DebugResponseWriterCloser) Flush() {
	// this is just a nop
}

// DumpRequest is a helper to dump a request in the same format that Close() dumps the response
func (w *DebugResponseWriterCloser) DumpRequest(r *http.Request) {
	b, _ := httputil.DumpRequest(r, true)
	log.Printf("---------- REQUEST ----------\n%s", b)
}

// DummyResponseWriterCloser implements only the context cancellation and otherwise does nothing.
type DummyResponseWriterCloser struct {
	http.ResponseWriter
	cancelFunc context.CancelFunc
}

func (w *DummyResponseWriterCloser) Write(p []byte) (int, error) {
	w.cancelFunc()
	return w.ResponseWriter.Write(p)
}

func (w *DummyResponseWriterCloser) WriteHeader(c int) {
	w.cancelFunc()
	w.ResponseWriter.WriteHeader(c)
}

func (w *DummyResponseWriterCloser) Close() error {
	w.cancelFunc()
	return nil
}

func (w *DummyResponseWriterCloser) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// GzipResponseWriterCloser implements gzip writing and context cancellation.
type GzipResponseWriterCloser struct {
	w             http.ResponseWriter
	gwriter       *gzip.Writer
	headerWritten bool
	cancelFunc    context.CancelFunc
}

func (w *GzipResponseWriterCloser) Header() http.Header {
	return w.w.Header()
}

func (w *GzipResponseWriterCloser) Write(p []byte) (int, error) {
	w.cancelFunc()
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.gwriter.Write(p)
}

func (w *GzipResponseWriterCloser) WriteHeader(c int) {
	w.cancelFunc()
	w.headerWritten = true
	w.w.WriteHeader(c)
}

func (w *GzipResponseWriterCloser) Close() error {
	w.cancelFunc()
	return w.gwriter.Close()
}

func (w *GzipResponseWriterCloser) Flush() {
	err := w.gwriter.Flush()
	if err != nil {
		log.Printf("gzip writer Flush() error: %v", err)
	}
	wf := w.w.(http.Flusher)
	wf.Flush()
}

// FIXME: figure out if we're going to keep the debug/dumping stuff in here or not - might be useful but needs think-through
// DebugCtxWrap is like GzipCtxWrap but instead of gzipping it implements debug logging that dumps full responses to the log.
// Use DumpRequest to dump the request in the same format.
func DebugCtxWrap(w http.ResponseWriter, r *http.Request) (*DebugResponseWriterCloser, context.Context) {

	ctx := r.Context()
	ctx, cfunc := context.WithCancel(ctx)

	wr := httptest.NewRecorder()

	ret := &DebugResponseWriterCloser{ResponseWriter: wr, origw: w, cancelFunc: cfunc}

	return ret, ctx

}

// GzipCtxWrap returns a ResponseWriterCloser that implements gzip response writing if possible
// and a context that is cancelled once the response header is written.  This makes it simple
// to ask the context if the request has been handled yet.
func GzipCtxWrap(w http.ResponseWriter, r *http.Request) (ResponseWriterCloser, context.Context) {

	ctx := r.Context()

	// prevent double wrap
	if w2, ok := w.(*GzipResponseWriterCloser); ok {
		return w2, ctx
	}

	ctx, cfunc := context.WithCancel(ctx)

	if !strings.Contains(r.Header.Get("accept-encoding"), "gzip") {
		return &DummyResponseWriterCloser{ResponseWriter: w, cancelFunc: cfunc}, ctx
	}

	gwriter := gzip.NewWriter(w)

	ret := &GzipResponseWriterCloser{w: w, gwriter: gwriter, cancelFunc: cfunc}

	w.Header().Set("content-encoding", "gzip")

	return ret, ctx
}

func NewGzipHandler() *GzipHandler {
	return &GzipHandler{}
}

type GzipHandler struct {
	NextHandler http.Handler
}

func (h *GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	newW, ctx := GzipCtxWrap(w, r)
	defer newW.Close() // make sure gzip stuff gets flushed properly

	r = r.WithContext(ctx)

	h.NextHandler.ServeHTTP(newW, r)
}

func (h *GzipHandler) SetNextHandler(next http.Handler) ChainedHandler {
	h.NextHandler = next
	return h
}

// Reads and reports an http error - does not expose anything to the outside
// world except a unique ID, which can be matched up with the appropriate log
// statement which has the details.
func HTTPError(w http.ResponseWriter, r *http.Request, err error, publicMessage string, code int) error {

	if err == nil {
		err = errors.New(publicMessage)
	}

	id := fmt.Sprintf("%x", time.Now().Unix()^mathrand.Int63())

	_, file, line, _ := runtime.Caller(1)

	w.Header().Set("x-error-id", id) // make a way for the client to programatically extract the error id
	http.Error(w, fmt.Sprintf("Error serving request (id=%q) %s", id, publicMessage), code)

	log.Printf("HTTPError: (id=%q) %s:%v | %v", id, file, line, err)

	return err
}

func NewCtxMapHandler(ctxMap map[string]interface{}) *CtxMapHandler {
	return &CtxMapHandler{
		CtxMap: ctxMap,
	}
}

// CtxMapHandler sets static items in the context during each request.
type CtxMapHandler struct {
	CtxMap      map[string]interface{}
	NextHandler http.Handler
}

func (h *CtxMapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	for k, v := range h.CtxMap {
		ctx = context.WithValue(ctx, k, v)
	}

	r = r.WithContext(ctx)

	h.NextHandler.ServeHTTP(w, r)
}

func (h *CtxMapHandler) SetNextHandler(next http.Handler) ChainedHandler {
	h.NextHandler = next
	return h
}
