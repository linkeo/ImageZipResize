package fileutil

import (
	"os"
	"path/filepath"
)

var cacheDir = ".resize.cache"

func GetCacheDir(base string) string {
	return filepath.Join(filepath.Clean(base), cacheDir)
}

func GetTempDir(base, filename string) (string, error) {
	rel, err := filepath.Rel(base, filename)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(GetCacheDir(base), rel)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", err
	}
	return dir, nil
}
