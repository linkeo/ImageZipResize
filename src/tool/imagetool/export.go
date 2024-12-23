package imagetool

import (
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"archive/zip"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const (
	backupDir = ".resize.backup"
	resizeTag = ".resized"

	extGIF       = ".gif"
	extJPEG      = ".jpeg"
	extJPEGAlias = ".jpg"
	extPNG       = ".png"
	extBMP       = ".bmp"
	extWEBP      = ".webp"
	extZip       = ".zip"
)

var exportLock sync.Mutex
var backupLock sync.Mutex
var supportedFileExt = map[string]bool{
	extGIF:       true,
	extJPEG:      true,
	extJPEGAlias: true,
	extPNG:       true,
	extBMP:       true,
	extWEBP:      true,
}

func IsSupportedImageFilename(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedFileExt[ext]
}

func IsResizedPath(filename string) bool {
	ext := filepath.Ext(filename)
	return strings.HasSuffix(filename, resizeTag+ext)
}

func IsOriginBackupPath(filename string) bool {
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
		return os.Rename(from, to+".backup")
	}
	return os.Rename(from, to)
}

func backupOrKeepOrigin(base, from string, to string) (float64, error) {
	fromStat, err := os.Stat(from)
	if err != nil {
		return 0, err
	}
	toStat, err := os.Stat(to)
	if err != nil {
		return 0, err
	}
	result := float64(toStat.Size()) / float64(fromStat.Size())
	//return result, nil
	if result < 1 {
		if err := backupOriginFile(base, from); err != nil {
			return 0, err
		}
		return result, nil
	}
	os.Remove(to)
	if err := os.Rename(from, getResizedName(from, path.Ext(from))); err != nil {
		return 0, err
	}
	return 1, nil
}

func writeResizedRGBImage(base, originFilename string, img image.Image) (float64, error) {
	toPath := getResizedName(originFilename, extJPEG)
	toFile, err := os.Create(toPath)
	if err != nil {
		return 0, err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	if err := jpeg.Encode(toFile, img, &jpeg.Options{Quality: 85}); err != nil {
		return 0, err
	}
	toFile.Close()
	return backupOrKeepOrigin(base, originFilename, toPath)
}

func writeResizedRGBAImage(base, originFilename string, img image.Image) (float64, error) {
	toPath := getResizedName(originFilename, extPNG)
	toFile, err := os.Create(toPath)
	if err != nil {
		return 0, err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	if err := png.Encode(toFile, img); err != nil {
		return 0, err
	}
	toFile.Close()
	return backupOrKeepOrigin(base, originFilename, toPath)
}

func writeResizedGIFImage(base, originFilename string, img *gif.GIF) (float64, error) {
	toPath := getResizedName(originFilename, extGIF)
	toFile, err := os.Create(toPath)
	if err != nil {
		return 0, err
	}
	defer toFile.Close()
	exportLock.Lock()
	defer exportLock.Unlock()
	if err := gif.EncodeAll(toFile, img); err != nil {
		return 0, err
	}
	toFile.Close()
	return backupOrKeepOrigin(base, originFilename, toPath)
}

//func writeResizedGIFImage(creator ImageCreator, img *gif.GIF) error {
//	file, err := creator()
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//	exportLock.Lock()
//	defer exportLock.Unlock()
//	return gif.EncodeAll(file, img)
//}

func newGIFWriter(img *gif.GIF) ImageWriter {
	return func(creator ImageCreator) error {
		file, err := creator()
		if err != nil {
			return err
		}
		defer file.Close()
		exportLock.Lock()
		defer exportLock.Unlock()
		return gif.EncodeAll(file, img)
	}
}

func isZipFile(filename string) bool {
	return filepath.Ext(filename) == extZip
}

func scanZipFile(filename string) (bool, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return false, err
	}
	defer reader.Close()
	for _, file := range reader.Reader.File {
		if IsSupportedImageFilename(file.Name) {
			return true, nil
		}
	}
	return false, nil
}

type ImageConverter func() ImageWriter
type ImageLoader func() (io.ReadCloser, error)
type ImageCreator func() (io.WriteCloser, error)
type ImageWriter func(creator ImageCreator) error
type ImageBackup func() error

func fileBackup(base string, filename string) ImageBackup {
	return func() error {
		return backupOriginFile(base, filename)
	}
}

func noopBackup() ImageBackup {
	return func() error {
		return nil
	}
}

func fileLoader(filename string) ImageLoader {
	return func() (io.ReadCloser, error) {
		return os.Open(filename)
	}
}

func fileCreator(filename string) ImageCreator {
	return func() (io.WriteCloser, error) {
		return os.Create(filename)
	}
}

func zipItemLoader(zipFile zip.File) ImageLoader {
	return func() (io.ReadCloser, error) {
		return zipFile.Open()
	}
}

func zipItemCreator(zipFile zip.Writer, filename string) ImageCreator {
	return func() (io.WriteCloser, error) {
		writer, err := zipFile.Create(filename)
		if err != nil {
			return nil, err
		}
		return fileutil.NewWriteCloser(writer), nil
	}
}
