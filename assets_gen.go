// +build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
)

func main() {
	config := vfsgen.Config{
		Input: assets,
		Tags:  "!dev",
	}

	err := vfsgen.Generate(config)
	if err != nil {
		log.Fatalln(err)
	}
}
