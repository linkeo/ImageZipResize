package imagetool

import (
	"errors"
	"image"
	"image/gif"
	"io"
)

func toPalettedImage(img image.Image) (*image.Paletted, error) {
	if img == nil {
		return nil, errors.New("cannot get paletted image from nil")
	}
	p, ok := img.(*image.Paletted)
	if ok {
		return p, nil
	}
	r, w := io.Pipe()
	ech := make(chan error, 1)
	go func() {
		ech <- gif.Encode(w, img, nil)
	}()
	result, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}
	if err := <-ech; err != nil {
		return nil, err
	}
	if len(result.Image) == 0 {
		return nil, errors.New("cannot read written gif image")
	}
	return result.Image[0], nil
}
