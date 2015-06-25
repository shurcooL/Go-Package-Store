// +build dev

package main

import (
	"net/http"

	"github.com/shurcooL/go/vfs/httpfs/union"
)

const production = false

var assets = union.New(map[string]http.FileSystem{
	"/assets": http.Dir("assets"),
})
