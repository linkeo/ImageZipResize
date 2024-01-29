package imagetool

import (
	"bufio"
	"bytes"
	"image"
	"image/gif"
	"os"
)

var gifMagic = []byte("GIF8?a")

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
	reader, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer reader.Close()
	buf := bufio.NewReader(reader)
	peek, err := buf.Peek(len(gifMagic))
	if err != nil {
		return false, err
	}
	return bytes.Equal(peek, gifMagic), nil
}
