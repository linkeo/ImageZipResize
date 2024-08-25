package flatten

import (
	"ImageZipResize/util/fileutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
)

var deflateCache map[string]bool
var deflateLock sync.Mutex

func Expand(dir string) (err error) {
	deflateLock.Lock()
	defer deflateLock.Unlock()
	deflateCache = make(map[string]bool)
	files, err := fileutil.ScanFiles(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if er := expand(dir, file); er != nil {
			err = multierror.Append(err, er)
		}
	}
	return err
}

func expand(dir string, file string) error {
	oldRel, err := filepath.Rel(dir, file)
	if err != nil {
		return err
	}
	newRel, changed := expandPath(oldRel)
	target := filepath.Join(dir, newRel)
	if !changed {
		log.Printf("not changed %s to %s", file, target)
		return nil
	}
	targetDir := filepath.Dir(target)
	if _, ok := deflateCache[targetDir]; !ok {
		err := os.MkdirAll(targetDir, 0777)
		if err != nil {
			return err
		}
		deflateCache[targetDir] = true
	}
	log.Printf("moving %s to %s", file, target)
	return os.Rename(file, target)
}
