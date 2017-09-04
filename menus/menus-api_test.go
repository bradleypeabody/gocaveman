package menus

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMenuAPIHandler(t *testing.T) {

	assert := assert.New(t)

	// fs := afero.NewBasePathFs(afero.NewOsFs(), "")
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/menu-data", 0775)

	fileMenus := NewFileMenus(fs, "/menu-data/")

	h := NewMenuAPIHandler(fileMenus)

	srv := httptest.NewServer(h)
	defer srv.Close()

	client := srv.Client()

	req, _ := http.NewRequest("POST", srv.URL+"/api/menus/test123", bytes.NewBufferString(`{"text":"Text Value","link":"Link Value"}`))
	req.Header.Set("content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(200, res.StatusCode)

	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("POST MENU RESPONSE:\n%s", b)

	// now retrieve it
	req, _ = http.NewRequest("GET", srv.URL+"/api/menus/test123", nil)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(200, res.StatusCode)

	b, err = httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GET MENU RESPONSE: %s", b)

	// get the list
	req, _ = http.NewRequest("GET", srv.URL+"/api/menus", nil)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(200, res.StatusCode)

	b, err = httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GET INDEX RESPONSE: %s", b)
	assert.Contains(string(b), "test123")

	// delete it
	req, _ = http.NewRequest("DELETE", srv.URL+"/api/menus/test123", nil)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(204, res.StatusCode)

	b, err = httputil.DumpResponse(res, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DELETE MENU RESPONSE: %s", b)
}
