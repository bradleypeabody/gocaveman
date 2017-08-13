package gocaveman

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func NewJSONFileMenus(fs afero.Fs, dir string) *JSONFileMenus {
	return &JSONFileMenus{Fs: fs, Dir: dir}
}

type JSONFileMenus struct {
	Fs  afero.Fs
	Dir string
}

func (f *JSONFileMenus) MenuIDs() ([]string, error) {

	dirf, err := f.Fs.Open(f.Dir)
	if err != nil {
		return nil, err
	}
	defer dirf.Close()

	fis, err := dirf.Readdir(-1)

	// fis, err := ioutil.ReadDir(f.Dir)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".json") {
			ret = append(ret, strings.TrimSuffix(fi.Name(), ".json"))
		}
	}
	return ret, nil
}

func (f *JSONFileMenus) ReadMenu(id string) (*Menu, error) {

	inf, err := f.Fs.Open(filepath.Join(f.Dir, id+".json"))
	if err != nil {
		return nil, err
	}
	defer inf.Close()

	b, err := ioutil.ReadAll(inf)

	// b, err := ioutil.ReadFile(filepath.Join(f.Dir, id+".json"))
	if err != nil {
		return nil, err
	}

	m := Menu{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (f *JSONFileMenus) DeleteMenu(id string) error {

	return f.Fs.Remove(filepath.Join(f.Dir, id+".json"))
}

func (f *JSONFileMenus) WriteMenu(id string, menu *Menu) error {

	outf, err := f.Fs.OpenFile(filepath.Join(f.Dir, id+".json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outf.Close()

	return json.NewEncoder(outf).Encode(menu)
}
