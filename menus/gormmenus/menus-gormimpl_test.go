package gormimpl

import (
	"testing"

	"github.com/bradleypeabody/gocaveman"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestGormMenu(t *testing.T) {

	db, err := gorm.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	defer db.Close()

	menus := NewGormMenus(db)
	db.AutoMigrate(menus.Models())
	MustCheckErrors(db)

	menu := &gocaveman.Menu{}
	menu.RootMenuItem = &gocaveman.MenuItem{
		Link: "/test1.html",
		Text: "Item 1",
	}

	err = menus.WriteMenu("test1", menu)
	if err != nil {
		t.Fatal(err)
	}

	menu, err = menus.ReadMenu("test1")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Menu: %+v", menu)

}
