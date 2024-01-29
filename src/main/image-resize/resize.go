package main

import (
	"ImageZipResize/tool/imagetool"
	"ImageZipResize/util/concurrent"
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
)

var resizeTarget = image.Pt(1600, 1600)

type entry struct {
	dir  string
	file string
}

func main() {
	files := slices.Filter(os.Args[1:], filters.PathIsRegularFile)
	dirs := slices.Filter(os.Args[1:], filters.PathIsDirectory)
	fmt.Println(os.Args, files, dirs)
	entries := make([]entry, 0)
	slices.ForEach(files, func(file string) {
		parent := filepath.Dir(file)
		entries = append(entries, entry{dir: parent, file: file})
	})
	slices.ForEach(dirs, func(dir string) {
		parent := filepath.Dir(dir)
		files, err := fileutil.ScanFiles(dir)
		if err != nil {
			log.Printf("scan files %s failed, %s", dir, err)
			return
		}
		slices.ForEach(files, func(file string) {
			entries = append(entries, entry{dir: parent, file: file})
		})
	})
	total := len(entries)
	curr := new(atomic.Int64)
	curr.Store(0)
	// update title by cmd := exec.Command("cmd", "/C", "title", "your_title_here")
	concurrent.ForEach(entries, func(en entry) {
		i := curr.Add(1)
		tag := fmt.Sprintf("%d/%d", i, total)
		resize(tag, en.dir, en.file)
	}, runtime.NumCPU()-1)
}

func resize(tag, base, file string) {
	log.Printf("[%s] %s resizing %s", tag, resizeTarget, file)
	err := imagetool.Resize(base, file, resizeTarget, imagetool.ModeInner.DoNotEnlarge())
	if err != nil {
		log.Printf("[%s] resize %s failed, %s", tag, file, err)
	}
}
