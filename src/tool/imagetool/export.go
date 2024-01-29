package imagetool

import (
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	backupDir = ".resize.backup"
	resizeTag = ".resized"

	extGIF  = ".gif"
	extJPEG = ".jpeg"
	extPNG  = ".png"
)

var exportLock sync.Mutex
var backupLock sync.Mutex

func isResizedPath(filename string) bool {
	ext := filepath.Ext(filename)
	return strings.HasSuffix(filename, resizeTag+ext)
}

func isOriginBackupPath(filename string) bool {
	paths := fileutil.SplitPath(filename)
	return slices.Any(paths, filters.Equal(backupDir))
}

func getResizedName(filename, newExt string) string {
	ext := filepath.Ext(filename)
	return filepath.Clean(strings.TrimSuffix(filename, ext) + resizeTag + newExt)
}

func getOriginOldPath(filename string) (string, error) {
	paths := fileutil.SplitPath(filename)
	_, index, found := slices.FindLast(paths, filters.Equal(backupDir))
	if !found {
		return "", errors.New("cannot get origin anchor")
	}
	newPaths := append(paths[:index], paths[index+1:]...)
	return filepath.Join(newPaths...), nil
}

func getOriginNewPath(base, filename string) (string, error) {
	rel, err := filepath.Rel(base, filename)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Clean(base), backupDir, rel), nil
}

func isFileExist(file string) (bool, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func backupOriginFile(base, from string) error {
	to, err := getOriginNewPath(base, from)
	if err != nil {
		return err
	}
	backupLock.Lock()
	defer backupLock.Unlock()
	if err := os.MkdirAll(filepath.Dir(to), 0777); err != nil {
		return err
	}
	isExist, err := isFileExist(to)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("origin file already exists")
	}
	return os.Rename(from, to)
}

func writeResizedRGBImage(base, originFilename string, img image.Image) error {
	if err := backupOriginFile(base, originFilename); err != nil {
		return err
	}
	toPath := getResizedName(originFilename, extJPEG)
	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	return jpeg.Encode(toFile, img, &jpeg.Options{Quality: 85})
}

func writeResizedRGBAImage(base, originFilename string, img image.Image) error {
	if err := backupOriginFile(base, originFilename); err != nil {
		return err
	}
	toPath := getResizedName(originFilename, extPNG)
	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	return png.Encode(toFile, img)
}

func writeResizedGIFImage(base string, originFilename string, img *gif.GIF) error {
	if err := backupOriginFile(base, originFilename); err != nil {
		return err
	}
	toPath := getResizedName(originFilename, extGIF)
	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	return gif.EncodeAll(toFile, img)
}
