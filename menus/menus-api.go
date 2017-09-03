package menus

import (
	"net/http"
	"path"
	"strings"

	cvm "github.com/bradleypeabody/gocaveman"
)

// TODO: use go generate to bring in menus.html as bin data

func NewMenuAPIHandler(menus MenusReaderWriter) *MenuAPIHandler {
	return &MenuAPIHandler{
		Menus:  menus, // FIXME: hm... should we not just get the menus from the context?? NO!
		Prefix: "/api/menus",
	}
}

type MenuAPIHandler struct {
	Menus  MenusReaderWriter
	Prefix string
}

func (h *MenuAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	p := path.Clean("/" + r.URL.Path)

	// ignore calls not meant for us
	if !(p == h.Prefix || strings.HasPrefix(p, h.Prefix+"/")) {
		return
	}

	if retErr := func() error {

		var pp cvm.PathParser

		switch {

		case r.Method == "GET" && p == h.Prefix:
			ids, err := h.Menus.MenuIDs()
			if err != nil {
				return err
			}

			ret := make([]map[string]interface{}, 0, len(ids))

			for _, id := range ids {
				ret = append(ret, map[string]interface{}{
					"id":   id,
					"path": h.Prefix + "/" + id,
				})
			}

			cvm.WriteJSON(w, ret, 200)
			break

		case (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") && pp.Match(p, h.Prefix+"/%S"):

			id := pp.ArgString(0)

			var menuItem MenuItem

			err := cvm.ReadJSON(r, &menuItem)
			if err != nil {
				return err
			}

			err = h.Menus.WriteMenu(id, &menuItem)
			if err != nil {
				return err
			}

			cvm.WriteJSON(w, menuItem, 200)

			break

			// case pp.Match(p, h.Prefix+"/%S"):

			// case pp.Match r.Method == "GET" && p == :

		case (r.Method == "GET" || r.Method == "HEAD") && pp.Match(p, h.Prefix+"/%S"):

			id := pp.ArgString(0)

			menuItem, err := h.Menus.ReadMenu(id)
			if err != nil {
				return err
			}
			if r.Method == "HEAD" {
				w.WriteHeader(200)
				break
			}

			cvm.WriteJSON(w, menuItem, 200)
			break

		case r.Method == "DELETE" && pp.Match(p, h.Prefix+"/%S"):

			id := pp.ArgString(0)

			err := h.Menus.DeleteMenu(id)
			if err != nil {
				return err
			}

			w.WriteHeader(204)
			break

		}

		return nil

	}(); retErr != nil {
		if retErr == ErrNotFound {
			http.NotFound(w, r)
		} else {
			cvm.HTTPError(w, r, retErr, "error during request processing", 500)
		}
	}

}
