package flatten

import (
	files2 "ImageZipResize/util/fileutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

func Flatten(dir string) (err error) {
	files, err := files2.ScanFiles(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if er := flatten(dir, file); er != nil {
			err = multierror.Append(err, er)
		}
	}
	return err
}

func flatten(dir string, file string) error {
	oldRel, err := filepath.Rel(dir, file)
	if err != nil {
		return err
	}
	newRel, changed := flattenPath(oldRel)
	target := filepath.Join(dir, newRel)
	if !changed {
		log.Printf("not changed %s to %s", file, target)
		return nil
	}
	log.Printf("moving %s to %s", file, target)
	return os.Rename(file, target)
}
