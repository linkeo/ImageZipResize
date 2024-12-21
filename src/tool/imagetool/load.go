package imagetool

import (
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

func init() {
	_ = webp.Decode
	_ = gif.Decode
	_ = tiff.Decode
	_ = jpeg.Decode
	_ = png.Decode
	_ = bmp.Decode
}

func loadImageConfig(filename string) (conf image.Config, format string, err error) {
	reader, err := os.Open(filename)
	if err != nil {
		return conf, format, err
	}
	defer reader.Close()
	return image.DecodeConfig(reader)
}

func loadStaticImage(filename string) (img image.Image, err error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	img, _, err = image.Decode(reader)
	return
}

func loadGifImage(filename string) (*gif.GIF, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return gif.DecodeAll(reader)
}

func isGifImage(filename string) (bool, error) {
	_, f, err := loadImageConfig(filename)
	if err != nil {
		return false, err
	}
	return f == "gif", nil
}

func IsImageFile(filename string) bool {
	_, _, err := loadImageConfig(filename)
	return err == nil
}
