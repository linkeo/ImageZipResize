package main

import (
	"ImageZipResize/util/concurrent"
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"
)

var datePrefixedPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\D`)

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
	files = slices.Filter(files, filters.Not(isDatePrefixed))
	total := len(files)
	curr := new(atomic.Int64)
	curr.Store(0)
	// update title by cmd := exec.Command("cmd", "/C", "title", "your_title_here")
	concurrent.ForEach(files, func(file string) {
		i := curr.Add(1)
		tag := fmt.Sprintf("%d/%d", i, total)
		rename(tag, file)
	}, runtime.NumCPU()-1)
}

func rename(tag string, path string) {
	target, err := toDatePrefixed(path)
	if err != nil {
		log.Printf("[%s] failed to get prefixed name, %s", tag, err)
		return
	}
	if err := os.Rename(path, target); err != nil {
		log.Printf("[%s] failed to rename %s to %s, %s", tag, path, target, err)
		return
	}
	log.Printf("[%s] rename %s to %s", tag, path, target)
}

func isDatePrefixed(path string) bool {
	name := filepath.Base(path)
	return datePrefixedPattern.MatchString(name)
}

func toDatePrefixed(path string) (string, error) {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	prefix := stat.ModTime().Format(time.DateOnly) + " "
	return filepath.Join(dir, prefix+name), nil
}
