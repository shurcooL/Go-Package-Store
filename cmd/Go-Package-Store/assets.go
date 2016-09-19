// +build dev

package main

import (
	"net/http"
	"path/filepath"

	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/octicons"
)

const production = false

var assets = union.New(map[string]http.FileSystem{
	"/assets":   gopherjs_http.NewFS(http.Dir(filepath.Join("..", "..", "assets"))),
	"/octicons": octicons.Assets,
})
