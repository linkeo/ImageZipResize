package imagetool

import (
	"archive/zip"
	"errors"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

type Mode struct {
	sizing      sizingMode
	noEnlarging bool
}

type sizingMode string

const (
	sizingOuter    sizingMode = "outer"
	sizingInner    sizingMode = "inner"
	sizingByWidth  sizingMode = "width"
	sizingByHeight sizingMode = "height"
)

var (
	ModeOuter  = Mode{sizing: sizingOuter}
	ModeInner  = Mode{sizing: sizingInner}
	ModeWidth  = Mode{sizing: sizingByWidth}
	ModeHeight = Mode{sizing: sizingByHeight}
)

func (m Mode) DoNotEnlarge() Mode {
	m.noEnlarging = true
	return m
}

func Resize(base string, filename string, to image.Point, mode Mode) error {
	if IsResizedPath(filename) {
		return nil
	}
	//if isZipFile(filename) {
	//	return resizeImagesInZip(base, filename, to, mode)
	//}
	isGif, err := isGifImage(filename)
	if err != nil {
		return err
	}
	if isGif {
		return resizeGIF(base, filename, to, mode)
	}
	return resizeStatic(base, filename, to, mode)
}

func resizeImagesInZip(base string, filename string, to image.Point, mode Mode) error {
	ok, err := scanZipFile(filename)
	if err != nil {
		return err
	}
	if !ok {
		// skip zip files without images
		return nil
	}
	src, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer src.Close()
	dstFile, err := os.Create(getResizedName(filename, extZip))
	if err != nil {
		return err
	}
	defer dstFile.Close()
	dst := zip.NewWriter(dstFile)
	defer dst.Close()
	for _, file := range src.Reader.File {
		if !IsSupportedImageFile(file.Name) {
			srcItem, err := file.OpenRaw()
			if err != nil {
				return err
			}
			dstItem, err := dst.CreateRaw(&file.FileHeader)
			if err != nil {
				return err
			}
			if _, err := io.Copy(dstItem, srcItem); err != nil {
				return err
			}
			continue
		}

	}
	return nil
}

func resizeGIF(base string, filename string, to image.Point, mode Mode) error {
	img, err := loadGifImage(filename)
	if err != nil {
		return err
	}
	for i, origin := range img.Image {
		result, err := resize(origin, to, mode)
		if err != nil {
			return err
		}
		img.Image[i], err = toPalettedImage(result)
		if err != nil {
			return err
		}
	}
	if err := backupOriginFile(base, filename); err != nil {
		return err
	}
	writer := newGIFWriter(img)
	creator := fileCreator(getResizedName(filename, extGIF))
	return writer(creator)
}

func resizeGif(reader io.Reader, to image.Point, mode Mode) (ImageWriter, error) {
	img, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, err
	}
	for i, origin := range img.Image {
		result, err := resize(origin, to, mode)
		if err != nil {
			return nil, err
		}
		img.Image[i], err = toPalettedImage(result)
		if err != nil {
			return nil, err
		}
	}
	return newGIFWriter(img), nil
}

func resizeStatic(base string, filename string, to image.Point, mode Mode) error {
	img, err := loadStaticImage(filename)
	if err != nil {
		return err
	}
	result, err := resize(img, to, mode)
	if err != nil {
		return err
	}
	if result.Opaque() {
		return writeResizedRGBImage(base, filename, result)
	}
	return writeResizedRGBAImage(base, filename, result)
}

func resize(img image.Image, desire image.Point, mode Mode) (*image.NRGBA, error) {
	from := img.Bounds().Size()
	//log.Printf("from=%s desire=%s", from, desire)
	target, err := getTargetSize(from, desire, mode)
	if err != nil {
		return nil, err
	}
	//log.Printf("from=%s desire=%s target=%s", from, desire, target)
	resized := imaging.Resize(img, target.X, target.Y, imaging.Lanczos)

	//log.Printf("from=%s to=%s", from, resized.Bounds().Size())
	return resized, err
}

func getTargetSize(from image.Point, to image.Point, mode Mode) (result image.Point, err error) {
	result = from
	if to.X <= 0 || to.Y <= 0 {
		err = errors.New("image cannot getTargetSize to zero")
		return
	}
	if from.X <= 0 || from.Y <= 0 {
		err = errors.New("image cannot getTargetSize from zero")
		return
	}
	var resultFloat FloatPoint
	resultFloat, err = getTargetSizeFloat(Float(from), Float(to), mode)
	if err != nil {
		return
	}
	result = resultFloat.ToPoint()
	return

}

func getTargetSizeFloat(from FloatPoint, to FloatPoint, mode Mode) (result FloatPoint, err error) {
	scalePoint := to.DivPoint(from)
	var scale float64
	switch mode.sizing {
	case sizingOuter:
		scale = max(scalePoint.X, scalePoint.Y)
	case sizingInner:
		scale = min(scalePoint.X, scalePoint.Y)
	case sizingByWidth:
		scale = scalePoint.X
	case sizingByHeight:
		scale = scalePoint.Y
	default:
		err = fmt.Errorf("unknown getTargetSize mode %s", mode.sizing)
		return
	}
	if mode.noEnlarging && scale > 1.0 {
		scale = 1.0
	}
	result = from.Mul(scale)
	return
}
