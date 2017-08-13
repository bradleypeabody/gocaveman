package gocaveman

type MenusReader interface {
	// FIXME: add context?????
	MenuIDs() ([]string, error)
	ReadMenu(id string) (*Menu, error)
}

type MenusWriter interface {
	// NewMenu(id string) (*Menu, error)
	DeleteMenu(id string) error
	WriteMenu(id string, menu *Menu) error
}

type MenusReaderWriter interface {
	MenusReader
	MenusWriter
}

// TODO: should we also provide an http.Handler that provides read-only access to menus as JSON?
// would allow templates to do more dynamic stuff is needed - probably we should have this

// FIXME: is there any point to Menu?  maybe just have MenuItem...
type Menu struct {
	// ID           string
	RootMenuItem *MenuItem
	Data         interface{}
}

// TODO: need to support hierarchy...

type MenuItem struct {
	Text     string
	Link     string
	Children []*MenuItem
	Parent   *MenuItem
	// TODO: should we also have Prev and Next???
	Data interface{} // FIXME: this probably should be map[string]interface{} to encourage best practices; i was thinking this could be a struct before but i think that's unlikely because it won't survive being marshalled both ways in JSON so we just pick something simple and workable, even if less flexible
}

func NewMenuItem(link string, text string) *MenuItem {
	return &MenuItem{Link: link, Text: text}
}
