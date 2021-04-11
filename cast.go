package main

import (
	"image"
	"image/color"
	"math"

	"github.com/unixpickle/model3d/model3d"
)

func CastImage(collider model3d.Collider, img image.Image, index, distance float64) *image.RGBA {
	bounds := img.Bounds()
	res := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rc, ok := collider.FirstRayCollision(&model3d.Ray{
				Origin:    model3d.XYZ(float64(x), float64(y), 100),
				Direction: model3d.Z(-1),
			})
			if !ok {
				continue
			}
			v1 := Refract(model3d.Z(1), rc.Normal, index)
			if math.IsNaN(v1.Norm()) {
				panic("internal reflection")
			}
			finalCoord := v1.Scale(distance / v1.Z).XY()
			finalX, finalY := int(finalCoord.X+float64(x)), int(finalCoord.Y+float64(y))
			res.Set(x, y, ReflectAt(img, finalX, finalY))
		}
	}

	return res
}

func Refract(rayDir, normal model3d.Coord3D, index float64) model3d.Coord3D {
	theta := math.Acos(math.Abs(rayDir.Dot(normal)))
	theta1 := math.Asin(math.Sin(theta) * Refraction)
	otherVec := rayDir.ProjectOut(normal)
	return normal.Scale(math.Cos(theta1)).Add(otherVec.Scale(math.Sin(theta1)))
}

func ReflectAt(img image.Image, x, y int) color.Color {
	bounds := img.Bounds()
	x = ReflectPad(bounds.Min.X, bounds.Max.X-1, x)
	y = ReflectPad(bounds.Min.Y, bounds.Max.Y-1, y)
	return img.At(x, y)
}

func ReflectPad(min, max, value int) int {
	for value < min || value > max {
		if value < min {
			value = min - value
		} else if value > max {
			value = 2*max - value
		}
	}
	return value
}
