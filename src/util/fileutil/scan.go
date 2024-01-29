package fileutil

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

var Separator = fmt.Sprintf("%c", filepath.Separator)

func ScanFiles(dir string) (files []string, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	return
}

func SplitPath(path string) []string {
	return strings.Split(path, Separator)
}
