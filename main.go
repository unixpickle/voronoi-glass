package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	ZScale        = 2.5
	Refraction    = 0.9
	ImageDistance = 100.0
)

func main() {
	img := ReadImage("example/landscape.png")
	bounds := img.Bounds()
	min := model2d.XY(float64(bounds.Min.X), float64(bounds.Min.Y))
	max := model2d.XY(float64(bounds.Max.X), float64(bounds.Max.Y))

	points := make([]model2d.Coord, 500)
	for i := range points {
		points[i] = model2d.NewCoordRandBounds(min, max)
	}
	log.Println("Creating Voronoi cells...")
	voronoi := VoronoiCells(min, max, points)
	log.Println("Repairing Voronoi cells...")
	voronoi.Repair(1e-8)

	// log.Println("Rendering Voronoi diagram...")
	// voronoi.Render("voronoi.png")

	log.Println("Creating Voronoi collider...")
	mesh := voronoi.Mesh()
	mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		return c.Add(model3d.Z(rand.NormFloat64() * ZScale))
	})
	collider := model3d.MeshToCollider(mesh)

	log.Println("Casting image...")
	res := CastImage(collider, img, Refraction, ImageDistance)
	f, err := os.Create("output.png")
	essentials.Must(err)
	defer f.Close()
	essentials.Must(png.Encode(f, res))
}

func ReadImage(path string) image.Image {
	f, err := os.Open(path)
	essentials.Must(err)
	defer f.Close()
	img, _, err := image.Decode(f)
	essentials.Must(err)
	return img
}
