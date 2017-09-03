package menus

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFileMenus(t *testing.T) {

	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	err := fs.Mkdir("/menus", 0755)
	if err != nil {
		t.Fatal(err)
	}

	fm := NewFileMenus(fs, "/menus")

	menu := NewMenuItem("/item1.html", "Item 1")

	err = fm.WriteMenu("test1", menu)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open("/menus/test1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DATA: %s", b)

	menu, err = fm.ReadMenu("test1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal("/item1.html", menu.Link)

}
