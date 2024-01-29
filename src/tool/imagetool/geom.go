package imagetool

import (
	"fmt"
	"image"
	"math"
)

type FloatPoint struct {
	X, Y float64
}

func (p FloatPoint) String() string {
	return fmt.Sprintf("(%.2f,%.2f)", p.X, p.Y)
}

func Float(point image.Point) FloatPoint {
	return FloatPoint{
		X: float64(point.X),
		Y: float64(point.Y),
	}
}

func (fp FloatPoint) ToPoint() image.Point {
	return image.Pt(int(math.Round(fp.X)), int(math.Round(fp.Y)))
}

func (fp FloatPoint) Mul(scale float64) FloatPoint {
	fp.X = fp.X * scale
	fp.Y = fp.Y * scale
	return fp
}

func (fp FloatPoint) MulPoint(scale FloatPoint) FloatPoint {
	fp.X = fp.X * scale.X
	fp.Y = fp.Y * scale.Y
	return fp
}

func (fp FloatPoint) DivPoint(scale FloatPoint) FloatPoint {
	fp.X = fp.X / scale.X
	fp.Y = fp.Y / scale.Y
	return fp
}
