package imagetool

import (
	"ImageZipResize/util/filters"
	"errors"
	"image"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var rollbackLock sync.Mutex

func Rollback(file string, desire image.Point, mode Mode) error {
	if !isOriginBackupPath(file) {
		return errors.New("rollback not resized origin image")
	}
	predictExt, err := predictResizedExt(file, desire, mode)
	if err != nil {
		return err
	}
	oldPath, err := getOriginOldPath(file)
	if err != nil {
		return err
	}
	rollbackLock.Lock()
	defer rollbackLock.Unlock()
	log.Printf("rollback %s to %s", file, oldPath)
	if err := os.Rename(file, oldPath); err != nil {
		return err
	}
	resized := getResizedName(oldPath, predictExt)
	if filters.PathIsRegularFile(resized) && filepath.Clean(resized) != filepath.Clean(oldPath) {
		log.Printf("remove resized file %s", resized)
		return os.Remove(resized)
	}
	return nil
}

func predictResizedExt(file string, desire image.Point, mode Mode) (string, error) {
	isGif, err := isGifImage(file)
	if err != nil {
		return "", err
	}
	if isGif {
		return extGIF, nil
	}
	img, err := loadStaticImage(file)
	if err != nil {
		return "", err
	}
	resized, err := resize(img, desire, mode)
	if err != nil {
		return "", err
	}
	if resized.Opaque() {
		return extJPEG, nil
	}
	return extPNG, nil
}
