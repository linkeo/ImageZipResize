package main

import (
	"ImageZipResize/tool/flatten"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"log"
	"os"
)

func main() {
	dirs := slices.Filter(os.Args[1:], filters.PathIsDirectory)
	for _, dir := range dirs {
		if err := flatten.Expand(dir); err != nil {
			log.Printf("flattening %q failed, %s", dir, err)
		}
	}
}
