package gormimpl

import (
	"bytes"
	"encoding/json"

	"github.com/bradleypeabody/gocaveman"
	"github.com/jinzhu/gorm"
)

func NewGormMenus(db *gorm.DB) *GormMenus {
	return &GormMenus{DB: db}
}

type GormMenus struct {
	DB *gorm.DB
}

func (gm *GormMenus) Models() []interface{} {
	return []interface{}{GormMenu{}}
}

type GormMenu struct {
	gorm.Model
	ID       string
	MenuJson string
}

func (g *GormMenus) MenuIDs() ([]string, error) {

	var gms []GormMenu

	// FIXME: we should not be selecting all fields here, would be faster not to
	g.DB.Find(&gms)
	err := CheckErrors(g.DB)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(gms))
	for _, gm := range gms {
		ids = append(ids, gm.ID)
	}

	return ids, nil
}

func (g *GormMenus) ReadMenu(id string) (*Menu, error) {

	gm := GormMenu{}
	g.DB.First(&gm, id)
	err := CheckErrors(g.DB)
	if err != nil {
		return nil, err
	}

	var menu gocaveman.Menu

	err = json.NewDecoder(bytes.NewReader([]byte(gm.MenuJson))).Decode(&menu)
	if err != nil {
		return nil, err
	}

	return &menu, nil
}

func (g *GormMenus) DeleteMenu(id string) error {

	g.DB.Delete(&GormMenu{ID: id})
	return CheckErrors(g.DB)
}

func (g *GormMenus) WriteMenu(id string, menu *Menu) error {

	b, err := json.Marshal(menu)
	if err != nil {
		return err
	}

	gormMenu := GormMenu{
		ID:       id,
		MenuJson: string(b),
	}

	g.DB.Save(&gormMenu)
	return CheckErrors(g.DB)
}
