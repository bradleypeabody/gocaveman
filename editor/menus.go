package editor

import (
	"net/http"

	"github.com/bradleypeabody/gocaveman"
)

// TODO: use go generate to bring in menus.html as bin data

func NewMenuHandler(menus gocaveman.MenusReaderWriter) *MenuHandler {
	return &MenuHandler{
		Menus:       menus, // FIXME: hm... should we not just get the menus from the context??
		BaseAPIPath: "/api/editor/menus",
	}
}

type MenuHandler struct {
	Menus       gocaveman.MenusReaderWriter
	BaseAPIPath string
}

func (h *MenuHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// FIXME: this is going to need security but we'll have to figure out how users first

	// TODO: serve api call(s) (REST style) - /api/menus/....

	if r.URL.Path == h.BaseAPIPath {

		// menu := h.Menus.ReadMenu(id)
		// write as JSON

	}

}
