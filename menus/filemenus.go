package menus

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func NewFileMenus(fs afero.Fs, dir string) *FileMenus {
	return &FileMenus{Fs: fs, Dir: dir}
}

type FileMenus struct {
	Fs  afero.Fs
	Dir string
}

func (f *FileMenus) MenuIDs() ([]string, error) {

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

func (f *FileMenus) ReadMenu(id string) (*MenuItem, error) {

	inf, err := f.Fs.Open(filepath.Join(f.Dir, id+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer inf.Close()

	b, err := ioutil.ReadAll(inf)

	// b, err := ioutil.ReadFile(filepath.Join(f.Dir, id+".json"))
	if err != nil {
		return nil, err
	}

	m := MenuItem{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (f *FileMenus) DeleteMenu(id string) error {

	return f.Fs.Remove(filepath.Join(f.Dir, id+".json"))
}

func (f *FileMenus) WriteMenu(id string, menu *MenuItem) error {

	outf, err := f.Fs.OpenFile(filepath.Join(f.Dir, id+".json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outf.Close()

	return json.NewEncoder(outf).Encode(menu)
}
