package utils

import (
	"math"
)

type Point struct {
	X, Y float64
}

func NewPoint(x, y float64) Point {
	return Point{x, y}
}

func Rotate(points []Point, rad float64) []Point {
	rotated := make([]Point, len(points))
	co := math.Cos(-rad)
	si := math.Sin(-rad)
	for i, point := range points {
		x1 := point.X*co - point.Y*si
		y1 := point.X*si + point.Y*co
		rotated[i] = NewPoint(x1, y1)
	}
	return rotated
}

func Tranlate(points []Point, offset *Point) []Point {
	translated := make([]Point, len(points))
	for i, point := range points {
		x1 := offset.X + point.X
		y1 := offset.Y + point.Y
		translated[i] = NewPoint(x1, y1)
	}
	return translated
}

func isBetween(x0, x, x1 float64) bool {
	return x >= x0 && x <= x1
}

func linesIntersect(x0, y0, x1, y1, a0, b0, a1, b1 float64) bool {
	// from https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
	// Ref Marina Gavrilova
	// four endpoints are x0, y0 & x1,y1 & a0,b0 & a1,b1
	// returned values xy and ab are the fractional distance along xy and ab
	// and are only defined when the result is true

	var xy, ab float64

	partial := false
	denom := (b0-b1)*(x0-x1) - (y0-y1)*(a0-a1)
	if denom == 0 {
		xy = -1.0
		ab = -1.0
	} else {
		xy = (a0*(y1-b1) + a1*(b0-y1) + x1*(b1-b0)) / denom
		partial = isBetween(0, xy, 1)
		if partial {
			// no point calculating this unless xy is between 0 & 1
			ab = (y1*(x0-a1) + b1*(x1-x0) + y0*(a1-x1)) / denom
		}
	}
	if partial && isBetween(0, ab, 1) {
		ab = 1 - ab
		xy = 1 - xy
		return true
	}
	return false
}

func LinesIntersect(l00, l01, l10, l11 *Point) bool {
	return linesIntersect(l00.X, l00.Y, l01.X, l01.Y, l10.X, l10.Y, l11.X, l11.Y)
}

func ShapeAndLineIntersect(shape []Point, l00, l01 *Point) bool {
	p := &shape[0]
	for i := 1; i < len(shape); i++ {
		if LinesIntersect(p, &shape[i], l00, l01) {
			return true
		}
		p = &shape[i]
	}
	if len(shape) > 1 && LinesIntersect(&shape[0], &shape[len(shape)-1], l00, l01) {
		return true
	}
	return false
}

func VectorLength(vec *Point) float64 {
	return VectorLengthXY(vec.X, vec.Y)
}

func VectorLengthXY(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

func LineLength(p0, p1 *Point) float64 {
	return VectorLengthXY(p1.X-p0.X, p1.Y-p0.Y)
}

func GetAngle(fromX, fromY, toX, toY float64) float64 {
	// this is the _visual_ angle in radians: 0 is right, PI/2 is up,
	// PI is left, and 3PI/2 is down.
	dx := toX - fromX
	dy := fromY - toY /* mathematical y direction: positive up. */
	var ret float64
	if dx == 0 {
		if dy >= 0 {
			ret = math.Pi / 2.0
		} else {
			ret = -math.Pi / 2.0
		}
	} else {
		ret = math.Atan(dy / dx)
	}
	/* find correct quadrant, and "normalize". atan returns the
	 * interval -PI/2 to PI/2, we want 0 to 2PI. */
	if ret >= 0.0 {
		if dy < 0 && dx < 0 {
			ret += math.Pi
		}
	} else {
		if dy > 0 && dx < 0 {
			ret += math.Pi
		} else {
			ret += 2.0 * math.Pi
		}
	}
	return ret
}
