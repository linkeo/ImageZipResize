package imagetool

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"reflect"
)

func toPalettedImage(img image.Image, p color.Palette) *image.Paletted {
	if p, ok := img.(*image.Paletted); ok {
		return p
	}
	paletted := image.NewPaletted(img.Bounds(), p)
	draw.Draw(paletted, img.Bounds(), img, img.Bounds().Min, draw.Src)
	//draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, img.Bounds().Min)
	return paletted
}

func copyRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func drawPaletted(dst *image.RGBA, img *image.Paletted) {
	draw.Draw(dst, img.Bounds(), img, img.Bounds().Min, draw.Over)
}

func resolveGifDisposals(img *gif.GIF) {
	canvasBounds := img.Image[0].Bounds()
	canvas := image.NewRGBA(canvasBounds)

	for index, frame := range img.Image {
		prev := copyRGBA(canvas)
		drawPaletted(canvas, frame)
		img.Image[index] = toPalettedImage(canvas, palette.Plan9)
		//img.Image[index] = toPalettedImage(canvas, frame.Palette)

		switch img.Disposal[index] {
		case gif.DisposalBackground:
			canvas = image.NewRGBA(canvasBounds)
		case gif.DisposalPrevious:
			canvas = prev
		case gif.DisposalNone:
		}
	}
}

func optimizeDisposalGif(img *gif.GIF) {
	for i := len(img.Image) - 1; i >= 2; i-- {
		prevBounds := differenceBounds(img.Image[i-1], img.Image[i])
		skipBounds := differenceBounds(img.Image[i-2], img.Image[i])
		if areaOfRectangle(skipBounds) < areaOfRectangle(prevBounds) {
			img.Disposal[i-1] = gif.DisposalPrevious
			img.Image[i] = img.Image[i].SubImage(skipBounds).(*image.Paletted)
		} else {
			img.Disposal[i-1] = gif.DisposalNone
			img.Image[i] = img.Image[i].SubImage(prevBounds).(*image.Paletted)
		}
	}
	prevBounds := differenceBounds(img.Image[1], img.Image[0])
	img.Disposal[0] = gif.DisposalNone
	img.Image[1].SubImage(prevBounds)

	for i, frame := range img.Image {
		old := len(frame.Palette)
		optimizePalette(frame)
		log.Printf("bounds %d: %v | %v %d%% palette: %d -> %d", i, img.Image[0].Bounds(), frame.Bounds(), 100*areaOfRectangle(frame.Bounds())/areaOfRectangle(img.Image[0].Bounds()), old, len(frame.Palette))
	}
}

func optimizePalette(img *image.Paletted) {
	mapping := make(map[uint8]uint8, len(img.Palette))
	var seq uint8 = 0
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			old := img.ColorIndexAt(x, y)
			if _, hit := mapping[old]; !hit {
				mapping[old] = seq
				seq++
			}
			img.SetColorIndex(x, y, mapping[old])
		}
	}
	if seq == 0 {
		mapping[0] = seq
		seq++
	}
	newPalette := make(color.Palette, seq)
	for o, n := range mapping {
		newPalette[n] = img.Palette[o]
	}
	img.Palette = newPalette
}

func areaOfRectangle(rect image.Rectangle) int {
	return rect.Dx() * rect.Dy()
}

func differenceBounds(base *image.Paletted, next *image.Paletted) image.Rectangle {
	rng := base.Bounds().Intersect(next.Bounds())
	if !reflect.DeepEqual(base.Palette, next.Palette) {
		return rng
	}
	ret := rng
	ret.Min, ret.Max = ret.Max, ret.Min
	//e := color.RGBA{R: 5, G: 5, B: 5, A: 10}
	for x := rng.Min.X; x < rng.Max.X; x++ {
		for y := rng.Min.Y; y < rng.Max.Y; y++ {
			//if !near(base.At(x, y), next.At(x, y), e) {
			if base.ColorIndexAt(x, y) != next.ColorIndexAt(x, y) {
				ret.Min.X = min(x, ret.Min.X)
				ret.Min.Y = min(y, ret.Min.Y)
				ret.Max.X = max(x, ret.Max.X)
				ret.Max.Y = max(y, ret.Max.Y)
			}
		}
	}
	return ret
}

func near(c1 color.Color, c2 color.Color, c3 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	r3, g3, b3, a3 := c3.RGBA()
	return (max(r1, r2)-min(r1, r2) < r3) && (max(g1, g2)-min(g1, g2) < g3) && (max(b1, b2)-min(b1, b2) < b3) && (max(a1, a2)-min(a1, a2) < a3)
}
