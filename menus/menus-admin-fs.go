package menus

import (
	"net/http"
	"path"
	"time"

	"github.com/bradleypeabody/gocaveman"
)

var ErrNotFound = gocaveman.ErrNotFound

var modTime = time.Now()

func NewAdminMenusViewFS() http.FileSystem {
	return gocaveman.NewHTTPFuncFS(func(name string) (http.File, error) {
		name = path.Clean("/" + name)
		if name == "/admin/menus/index.gohtml" {
			return gocaveman.NewHTTPBytesFile(name, modTime, adminMenusListViewBytes), nil
		}
		if name == "/admin/menus/edit.gohtml" {
			return gocaveman.NewHTTPBytesFile(name, modTime, adminMenusEditViewBytes), nil
		}
		return nil, ErrNotFound
	})
}

var adminMenusListViewBytes = []byte(`{{template "/admin-page.gohtml" .}}
{{define "body"}}
List view
{{end}}
`)

var adminMenusEditViewBytes = []byte(`{{template "/admin-page.gohtml" .}}
{{define "body"}}
Edit view
{{end}}
`)
