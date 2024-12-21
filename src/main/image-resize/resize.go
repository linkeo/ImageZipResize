package main

import (
	"ImageZipResize/tool/imagetool"
	"ImageZipResize/util/concurrent"
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"ImageZipResize/util/system"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	slices_ "slices"
	"strings"
	"sync/atomic"
	"time"
)

var resizeTarget = image.Pt(1440, 1440)

type entry struct {
	root    string
	file    string
	isCover bool
}

func collect(files []string, dirs []string) ([]entry, map[string]struct{}) {
	entries := make([]entry, 0)
	roots := make(map[string]struct{})
	slices.ForEach(files, func(file string) {
		root := filepath.Dir(file)
		roots[root] = struct{}{}
		entries = append(entries, entry{root: root, file: file})
	})
	slices.ForEach(dirs, func(dir string) {
		root := filepath.Dir(dir)
		roots[root] = struct{}{}
		found, err := fileutil.ScanFiles(dir)
		if err != nil {
			log.Printf("scan files %s failed, %s", dir, err)
			return
		}
		slices.ForEach(found, func(file string) {
			entries = append(entries, entry{root: root, file: file})
		})
	})
	return entries, roots
}

func arrange(entries []entry) []entry {
	entries = slices.Filter(entries, func(e entry) bool {
		if !imagetool.IsSupportedImageFilename(e.file) {
			return false
		}
		if imagetool.IsOriginBackupPath(e.file) {
			return false
		}
		if imagetool.IsResizedPath(e.file) {
			return false
		}
		return true
	})
	slices_.SortStableFunc(entries, func(left entry, right entry) int {
		return strings.Compare(left.file, right.file)
	})
	for i := range entries {
		if i == 0 || filepath.Dir(entries[i].file) != filepath.Dir(entries[i-1].file) {
			entries[i].isCover = true
		}
	}
	return entries
}

func main() {
	args := slices.Filter(os.Args[1:], func(value string) bool {
		return !strings.HasPrefix(value, "--")
	})
	files := slices.Filter(args, filters.PathIsRegularFile)
	dirs := slices.Filter(args, filters.PathIsDirectory)
	for _, dir := range dirs {
		log.Printf("resize directory argument: %s", dir)
	}
	for _, file := range files {
		log.Printf("resize file argument: %s", file)
	}
	entries, roots := collect(files, dirs)
	entries = arrange(entries)

	total := len(entries)
	todoAtomic := new(atomic.Int64)
	todoAtomic.Store(int64(total))
	indexAtomic := new(atomic.Int64)
	indexAtomic.Store(0)
	doneAtomic := new(atomic.Int64)
	doneAtomic.Store(0)
	// update title by cmd := exec.Command("cmd", "/C", "title", "your_title_here")
	width := len(fmt.Sprintf("%d", total))
	par := system.GetParallelism()
	mem := system.GetMemoryLimit()
	log.Printf("resize %d images with parallelism %d, each with %s memory limit.", total, par, mem)
	startTime := time.Now()
	eta := func(delta int64) time.Duration {
		dn := doneAtomic.Add(delta)
		if dn == 0 {
			return time.Minute
		}
		td := todoAtomic.Add(-1)
		now := time.Now()
		elapsed := now.Sub(startTime)
		left := elapsed * time.Duration(td) / time.Duration(dn)
		return left.Truncate(time.Second)
	}
	defer func() {
		for root := range roots {
			os.RemoveAll(fileutil.GetCacheDir(root))
		}
	}()
	concurrent.ForEach(entries, func(en entry) {
		i := indexAtomic.Add(1)
		tag := fmt.Sprintf("%*d/%d", width, i, total)
		resize(tag, en, eta)
	}, int(par))
}

func resize(tag string, en entry, eta func(int64) time.Duration) {
	result, err := imagetool.Resize(en.root, en.file, en.isCover, resizeTarget, imagetool.ModeContain.DoNotEnlarge())
	if err != nil {
		log.Printf("[%s] %s resize failed, %s, %s, ETA %s", tag, resizeTarget, en.file, err, eta(0))
	} else {
		log.Printf("[%s] %s resize %7s, %s, ETA %s", tag, resizeTarget, compressRate(result), en.file, eta(1))
	}
}

func compressRate(rate float64) string {
	deltaPercent := (rate - 1) * 100
	if deltaPercent == 0.0 {
		return "-0%"
	}
	if deltaPercent < 0.0 {
		return fmt.Sprintf("%.2f%%", deltaPercent)
	}
	return fmt.Sprintf("+%.2f%%", deltaPercent)
}
