package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var numPoints int
	var noiseLevel float64
	var refraction float64
	var imageDist float64
	var inPath string
	var outPath string
	var useNN bool

	flag.IntVar(&numPoints, "points", 500, "points in Voronoi diagram")
	flag.Float64Var(&noiseLevel, "noise", 0.5, "scale of Z-axis noise")
	flag.Float64Var(&refraction, "refraction", 0.7, "index of refraction")
	flag.Float64Var(&imageDist, "image-dist", 100.0, "effective distance of photo from screen")
	flag.StringVar(&inPath, "in-path", "example/landscape.jpg", "input image")
	flag.StringVar(&outPath, "out-path", "output.png", "output image")
	flag.BoolVar(&useNN, "use-nn", false, "use nearest neighbors instead of a mesh")
	flag.Parse()

	img := ReadImage(inPath)
	bounds := img.Bounds()
	min := model2d.XY(float64(bounds.Min.X), float64(bounds.Min.Y))
	max := model2d.XY(float64(bounds.Max.X), float64(bounds.Max.Y))

	points := make([]model2d.Coord, numPoints)
	for i := range points {
		points[i] = model2d.NewCoordRandBounds(min, max)
	}

	var res image.Image
	if useNN {
		log.Println("Generating normals...")
		normals := make([]model3d.Coord3D, len(points))
		for i := range normals {
			noise := model3d.XY(rand.NormFloat64(), rand.NormFloat64())
			normals[i] = model3d.Z(1).Add(noise.Scale(noiseLevel)).Normalize()
		}
		log.Println("Casting image...")
		res = CastImageNN(points, normals, img, refraction, imageDist)
	} else {
		log.Println("Creating Voronoi cells...")
		voronoi := VoronoiCells(min, max, points)
		log.Println("Repairing Voronoi cells...")
		voronoi.Repair(1e-8)

		log.Println("Creating Voronoi collider...")
		mesh := voronoi.Mesh()
		mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
			sensitivity := NormalSensitivity(mesh, c)
			return c.Add(model3d.Z(2 * (rand.Float64() - 0.5) / sensitivity * noiseLevel))
		})
		collider := model3d.MeshToCollider(mesh)

		log.Println("Casting image...")
		res = CastImage(collider, img, refraction, imageDist)
	}

	f, err := os.Create(outPath)
	essentials.Must(err)
	defer f.Close()
	essentials.Must(png.Encode(f, res))
}

func NormalSensitivity(m *model3d.Mesh, c model3d.Coord3D) float64 {
	max := 0.0
	for _, t := range m.Find(c) {
		max = math.Max(max, TriangleSensitivity(t, c))
	}
	return max
}

func TriangleSensitivity(t *model3d.Triangle, c model3d.Coord3D) float64 {
	for i, c1 := range t {
		if c1 == c {
			oldNormal := t.Normal().XY().Norm()
			t[i].Z += 1e-5
			newNormal := t.Normal().XY().Norm()
			t[i] = c
			return (newNormal - oldNormal) / 1e-5
		}
	}
	panic("unreachable")
}

func ReadImage(path string) image.Image {
	f, err := os.Open(path)
	essentials.Must(err)
	defer f.Close()
	img, _, err := image.Decode(f)
	essentials.Must(err)
	return img
}
