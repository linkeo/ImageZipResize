package imagetool

import (
	"bufio"
	"bytes"
	"image"
	"image/gif"
	"os"
)

var gifVersions = [][]byte{[]byte("GIF87a"), []byte("GIF89a")}

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
	peek, err := buf.Peek(len(gifVersions[0]))
	if err != nil {
		return false, err
	}

	for _, ver := range gifVersions {
		if bytes.Equal(peek, ver) {
			return true, nil
		}
	}
	return false, nil
}
