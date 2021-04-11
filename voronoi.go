package main

import (
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

type VoronoiCell struct {
	Center model2d.Coord
	Edges  []*model2d.Segment
}

// VoronoiCells computes the voronoi cells for a list of
// coordinates, assuming they are all contained within a
// bounding box.
//
// The resulting Voronoi cells may be slightly misaligned,
// i.e. adjacent edges' coordinates may differ due to
// rounding errors. See RepairVoronoiCells().
func VoronoiCells(min, max model2d.Coord, coords []model2d.Coord) []*VoronoiCell {
	cells := make([]*VoronoiCell, len(coords))
	for i, c := range coords {
		collection := newFilteredCoords(min, max, coords)
		for len(collection.Points) > 1 {
			neighbor := collection.NearestNeighbor(c)
			normal := neighbor.Sub(c).Normalize()
			constraint := &model2d.LinearConstraint{
				Normal: normal,
				Max:    normal.Dot(neighbor.Mid(c)),
			}
			collection.Constrain(constraint)
		}
		cells[i] = collection.Cell(c)
	}
	return cells
}

// RepairVoronoiCells merges nearly identical coordinates
// in a Voronoi diagram to make a well-connected diagram.
func RepairVoronoiCells(cells []*VoronoiCell, epsilon float64) {
	coordSet := map[model2d.Coord]bool{}
	coordSlice := []model2d.Coord{}
	for _, cell := range cells {
		for _, s := range cell.Edges {
			for _, p := range s {
				if !coordSet[p] {
					coordSet[p] = true
					coordSlice = append(coordSlice, p)
				}
			}
		}
	}
	tree := model2d.NewCoordTree(coordSlice)

	mapping := map[model2d.Coord]model2d.Coord{}
	for _, c := range coordSlice {
		if !coordSet[c] {
			continue
		}
		neighbors := neighborsInDistance(tree, c, epsilon)
		for _, n := range neighbors {
			if coordSet[n] {
				coordSet[n] = false
				mapping[n] = c
			}
		}
	}

	for _, cell := range cells {
		for i := 0; i < len(cell.Edges); i++ {
			edge := cell.Edges[i]
			for j, c := range edge {
				edge[j] = mapping[c]
			}
			if edge[0] == edge[1] {
				// This was almost a singular edge.
				essentials.UnorderedDelete(&cell.Edges, i)
			}
		}
	}
}

func neighborsInDistance(tree *model2d.CoordTree, c model2d.Coord, epsilon float64) []model2d.Coord {
	for k := 2; true; k++ {
		neighbors := tree.KNN(k, c)
		if len(neighbors) < k {
			return neighbors
		}
		if neighbors[len(neighbors)-1].Dist(c) > epsilon {
			return neighbors[:len(neighbors)-1]
		}
	}
	panic("unreachable")
}

type filteredCoords struct {
	Constraints model2d.ConvexPolytope
	Points      []model2d.Coord
}

func newFilteredCoords(min, max model2d.Coord, coords []model2d.Coord) *filteredCoords {
	res := &filteredCoords{
		Constraints: model2d.NewConvexPolytopeRect(min, max),
		Points:      make([]model2d.Coord, 0, len(coords)),
	}
	for _, c := range coords {
		if res.Constraints.Contains(c) {
			res.Points = append(res.Points, c)
		}
	}
	return res
}

func (f *filteredCoords) Constrain(l *model2d.LinearConstraint) {
	f.Constraints = append(f.Constraints, l)
	for i := 0; i < len(f.Points); i++ {
		if !l.Contains(f.Points[i]) {
			f.Points[i] = f.Points[len(f.Points)-1]
			f.Points = f.Points[:len(f.Points)-1]
		}
	}
}

func (f *filteredCoords) NearestNeighbor(c model2d.Coord) model2d.Coord {
	minDist := math.Inf(1)
	minPoint := c
	for _, p := range f.Points {
		if c == p {
			continue
		}
		dist := c.SquaredDist(p)
		if dist < minDist {
			minDist = dist
			minPoint = p
		}
	}
	return minPoint
}

func (f *filteredCoords) Cell(center model2d.Coord) *VoronoiCell {
	return &VoronoiCell{
		Center: center,
		Edges:  f.Constraints.Mesh().SegmentSlice(),
	}
}
