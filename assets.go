// +build dev

package main

import "net/http"

const production = false

var assets = http.Dir("./assets/")
