package imagetool

import (
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/system"
	"archive/zip"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path"
)

type Mode struct {
	sizing      sizingMode
	noEnlarging bool
}

type sizingMode string

const (
	sizingStretch  sizingMode = "stretch"
	sizingFill     sizingMode = "fill"
	sizingContain  sizingMode = "contain"
	sizingCover    sizingMode = "cover"
	sizingByWidth  sizingMode = "width"
	sizingByHeight sizingMode = "height"
)

var (
	ModeStretch = Mode{sizing: sizingStretch}
	ModeFill    = Mode{sizing: sizingFill}
	ModeContain = Mode{sizing: sizingContain}
	ModeCover   = Mode{sizing: sizingCover}
	ModeWidth   = Mode{sizing: sizingByWidth}
	ModeHeight  = Mode{sizing: sizingByHeight}
)

func (m Mode) DoNotEnlarge() Mode {
	m.noEnlarging = true
	return m
}

func Resize(base string, filename string, isCover bool, to image.Point, mode Mode) (float64, error) {
	if IsResizedPath(filename) {
		return 0, errors.New("file is already resized")
	}
	//if isZipFile(filename) {
	//	return resizeImagesInZip(base, filename, to, mode)
	//}
	if !IsImageFile(filename) {
		return 0, errors.New("file is not an image")
	}
	return resizeMagick(base, filename, isCover, to, mode)
	//isGif, err := isGifImage(filename)
	//if err != nil {
	//	return 0, err
	//}
	//if isGif {
	//	return resizeGIF(base, filename, to, mode)
	//}
	//return resizeStatic(base, filename, to, mode)
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
		if !IsSupportedImageFilename(file.Name) {
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

func resizeGIF(base string, filename string, to image.Point, mode Mode) (float64, error) {
	img, err := loadGifImage(filename)
	if err != nil {
		return 0, err
	}
	resolveGifDisposals(img)
	for i, origin := range img.Image {
		result, err := resize(origin, to, mode)
		if err != nil {
			return 0, err
		}
		img.Image[i] = toPalettedImage(result, origin.Palette)
	}
	optimizeDisposalGif(img)
	return writeResizedGIFImage(base, filename, img)
	//if err := backupOriginFile(base, filename); err != nil {
	//	return err
	//}
	//writer := newGIFWriter(img)
	//creator := fileCreator(getResizedName(filename, extGIF))
	//return writer(creator)
}

func resizeGif(reader io.Reader, to image.Point, mode Mode) (ImageWriter, error) {
	img, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, err
	}
	resolveGifDisposals(img)
	for i, origin := range img.Image {
		result, err := resize(origin, to, mode)
		if err != nil {
			return nil, err
		}
		img.Image[i] = toPalettedImage(result, origin.Palette)
	}
	optimizeDisposalGif(img)
	return newGIFWriter(img), nil
}

func resizeMagick(base string, filename string, isCover bool, to image.Point, mode Mode) (float64, error) {
	ext := extWEBP
	if isCover {
		ext = path.Ext(filename)
	}
	toPath := getResizedName(filename, ext)
	cmd := exec.Command("magick", filename, "-strip", "-coalesce", "-resize", magickResizeOption(to, mode), "-quality", "90", "-define", "webp:near-lossless=90", toPath)
	tmp, err := fileutil.GetTempDir(base, filename)
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tmp)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("MAGICK_MEMORY_LIMIT=%s", system.GetMemoryLimit()),
		fmt.Sprintf("MAGICK_MAP_LIMIT=%s", system.GetMemoryLimit()),
		fmt.Sprintf("MAGICK_DISK_LIMIT=%s", system.GetMemoryLimit()),
		fmt.Sprintf("MAGICK_TEMPORARY_PATH=%s", tmp))
	//log.Printf("command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		os.Remove(toPath)
		return 0, err
	}
	return backupOrKeepOrigin(base, filename, toPath)
}

func magickResizeOption(size image.Point, mode Mode) string {
	switch mode.sizing {
	case sizingContain:
		if mode.noEnlarging {
			return fmt.Sprintf("%dx%d>", size.X, size.Y)
		}
		return fmt.Sprintf("%dx%d", size.X, size.Y)
	case sizingByWidth:
		if mode.noEnlarging {
			return fmt.Sprintf("%dx>", size.X)
		}
		return fmt.Sprintf("%dx", size.X)
	case sizingByHeight:
		if mode.noEnlarging {
			return fmt.Sprintf("x%d>", size.Y)
		}
		return fmt.Sprintf("x%d", size.Y)
	case sizingStretch:
		return fmt.Sprintf("%dx%d!", size.X, size.Y)
	case sizingFill:
		return fmt.Sprintf("%dx%d^", size.X, size.Y)
	case sizingCover:
		return fmt.Sprintf("%dx%d<", size.X, size.Y)
	}
	return fmt.Sprintf("%dx%d", size.X, size.Y)
}

func resizeStatic(base string, filename string, to image.Point, mode Mode) (float64, error) {
	img, err := loadStaticImage(filename)
	if err != nil {
		return 0, err
	}
	result, err := resize(img, to, mode)
	if err != nil {
		return 0, err
	}
	if almostOpaque(result) {
		return writeResizedRGBImage(base, filename, result)
	}
	return writeResizedRGBAImage(base, filename, result)
}

func almostOpaque(p image.Image) bool {
	if p.Bounds().Empty() {
		return true
	}
	for x := p.Bounds().Min.X; x < p.Bounds().Max.X; x++ {
		for y := p.Bounds().Min.Y; y < p.Bounds().Max.Y; y++ {
			c := p.At(x, y)
			_, _, _, a := c.RGBA()
			if a <= 0xf0 {
				return false
			}
		}
	}
	return true
}

func resize(img image.Image, desire image.Point, mode Mode) (image.Image, error) {
	from := img.Bounds().Size()
	//log.Printf("from=%s desire=%s", from, desire)
	target, err := getTargetSize(from, desire, mode)
	if from == target {
		return img, nil
	}
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
	case sizingCover:
		scale = max(scalePoint.X, scalePoint.Y)
	case sizingContain:
		scale = min(scalePoint.X, scalePoint.Y)
	case sizingByWidth:
		scale = scalePoint.X
	case sizingByHeight:
		scale = scalePoint.Y
	case sizingStretch:
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
