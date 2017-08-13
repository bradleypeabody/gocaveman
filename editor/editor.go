package editor

import (
	"net/http"

	"github.com/bradleypeabody/gocaveman"
)

// TODO: do go generate for: bin/go-bindata -pkg editor -o src/github.com/bradleypeabody/gocaveman/editor/bindata.go -prefix src/github.com/bradleypeabody/gocaveman -ignore '[.]go$' src/github.com/bradleypeabody/gocaveman/editor/

//OLD go:generate go-bindata -pkg editor -o src/github.com/bradleypeabody/gocaveman/editor/bindata.go -prefix src/github.com/bradleypeabody/gocaveman -ignore '[.]go$' src/github.com/bradleypeabody/gocaveman/editor/
//go:generate go-bindata -pkg editor -o bindata.go -ignore .go$ ./

// NewDefaultPagesFs adapts the static assets in the editor package to a usable filesystem.
func NewDefaultPagesFs() http.FileSystem {

	return &gocaveman.HTTPBindataFs{
		AssetFunc:     Asset,
		AssetDirFunc:  AssetDir,
		AssetInfoFunc: AssetInfo,
		Prepend:       "editor/",
	}
}
