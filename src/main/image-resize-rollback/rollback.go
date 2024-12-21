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
	"runtime"
	"sync/atomic"
)

var resizeTarget = image.Pt(1600, 1600)

func main() {
	args := os.Args[1:]
	files := slices.Filter(args, filters.PathIsRegularFile)
	dirs := slices.Filter(args, filters.PathIsDirectory)
	fmt.Println(os.Args, files, dirs)
	slices.ForEach(dirs, func(dir string) {
		found, err := fileutil.ScanFiles(dir)
		if err != nil {
			log.Printf("scan files %s failed, %s", dir, err)
			return
		}
		slices.ForEach(found, func(file string) {
			files = append(files, file)
		})
	})
	files = slices.Filter(files, imagetool.IsSupportedImageFilename)
	files = slices.Filter(files, imagetool.IsOriginBackupPath)
	total := len(files)
	curr := new(atomic.Int64)
	curr.Store(0)
	// update title by cmd := exec.Command("cmd", "/C", "title", "your_title_here")
	concurrent.ForEach(files, func(file string) {
		i := curr.Add(1)
		tag := fmt.Sprintf("%d/%d", i, total)
		rollback(tag, file)
	}, runtime.NumCPU()-1)
}

func rollback(tag, file string) {
	log.Printf("[%s] rollback %s", tag, file)
	err := imagetool.Rollback(file, resizeTarget, imagetool.ModeContain.DoNotEnlarge())
	if err != nil {
		log.Printf("[%s] rollback %s failed, %s", tag, file, err)
	}
}
