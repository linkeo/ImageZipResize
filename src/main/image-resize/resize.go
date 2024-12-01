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

var resizeTarget = image.Pt(1440, 1440)

type entry struct {
	dir  string
	file string
}

func main() {
	args := os.Args[1:]
	files := slices.Filter(args, filters.PathIsRegularFile)
	dirs := slices.Filter(args, filters.PathIsDirectory)
	fmt.Println(os.Args, files, dirs)
	entries := make([]entry, 0)
	slices.ForEach(files, func(file string) {
		parent := filepath.Dir(file)
		entries = append(entries, entry{dir: parent, file: file})
	})
	slices.ForEach(dirs, func(dir string) {
		parent := filepath.Dir(dir)
		found, err := fileutil.ScanFiles(dir)
		if err != nil {
			log.Printf("scan files %s failed, %s", dir, err)
			return
		}
		slices.ForEach(found, func(file string) {
			entries = append(entries, entry{dir: parent, file: file})
		})
	})
	files = slices.Filter(files, imagetool.IsSupportedImageFile)
	files = slices.Filter(files, filters.Not(imagetool.IsOriginBackupPath))
	files = slices.Filter(files, filters.Not(imagetool.IsResizedPath))
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
	result, err := imagetool.Resize(base, file, resizeTarget, imagetool.ModeOuter.DoNotEnlarge())
	if err != nil {
		log.Printf("[%s] %s resize %s failed, %s", tag, resizeTarget, file, err)
	} else {
		log.Printf("[%s] %s resize %s succeeded. %s", tag, resizeTarget, file, compressRate(result))
	}
}

func compressRate(rate float64) string {
	deltaPercent := (rate - 1) * 100
	if deltaPercent == 0.0 {
		return "-"
	}
	if deltaPercent < 0.0 {
		return fmt.Sprintf("%.2f%%", deltaPercent)
	}
	return fmt.Sprintf("+%.2f%%", deltaPercent)
}
