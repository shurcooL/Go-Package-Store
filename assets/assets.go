//go:build dev

package assets

import (
	"go/build"
	"log"
	"net/http"

	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/httpfs/union"
)

// Assets contains assets for Go Package Store.
var Assets = union.New(map[string]http.FileSystem{
	"/assets":      gopherjs_http.NewFS(http.Dir(importPathToDir("github.com/shurcooL/Go-Package-Store/_data"))),
	"/frontend.js": gopherjs_http.Package("github.com/shurcooL/Go-Package-Store/frontend"),
})

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}
