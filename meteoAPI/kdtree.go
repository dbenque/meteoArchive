package meteoAPI

import "github.com/biogo/store/kdtree"

//import "code.google.com/p/biogo.store/kdtree"

// ---------------Implementation of kdtree point interface for station --------------

// Compare compare stations on a given dimension
func (p Station) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(Station)
	return p.getCoord()[d] - q.getCoord()[d]
}

//Dims Return the number of dimension associated to a station
func (p Station) Dims() int { return coordDim }

//Distance compute the distance between 2 stations
func (p Station) Distance(c kdtree.Comparable) float64 {
	q := c.(Station)
	var sum float64
	for dim, c := range p.getCoord() {
		d := c - q.getCoord()[dim]
		sum += d * d
	}
	return sum
}

//Index return the station associated to the given index
func (p Stations) Index(i int) kdtree.Comparable { return p[i] }

//Len return the number of stations stored in the list of stations
func (p Stations) Len() int { return len(p) }

//Pivot compute the pivot index on the given dimension
func (p Stations) Pivot(d kdtree.Dim) int { return stPlane{Stations: p, Dim: d}.Pivot() }

//Slice slicer
func (p Stations) Slice(start, end int) kdtree.Interface { return p[start:end] }

// func (p Stations) Bounds() *kdtree.Bounding {
// 	s := [2]Station{}
// 	s[0].coord[0] = -10000
// 	s[0].coord[1] = -10000
// 	s[0].coord[2] = -10000
// 	s[0].coordCached = true
// 	s[1].coord[0] = 10000
// 	s[1].coord[1] = 10000
// 	s[1].coord[2] = 10000
// 	s[1].coordCached = true
//
// 	b := new(kdtree.Bounding)
// 	b[0] = s[0]
// 	b[1] = s[1]
// 	return b
//
// }

// An nbPlane is a wrapping type that allows a Points type be pivoted on a dimension.
type stPlane struct {
	kdtree.Dim
	Stations
}

func (p stPlane) Less(i, j int) bool {
	return p.Stations[i].getCoord()[p.Dim] < p.Stations[j].getCoord()[p.Dim]
}

func medianOf(list kdtree.SortSlicer) int {
	n := list.Len()
	kdtree.Select(list.Slice(0, n), n/2)
	return n / 2
}

//Pivot compute the pivot for the given plane
func (p stPlane) Pivot() int { return kdtree.Partition(p, medianOf(p)) }

// const (
// 	// Randoms is the maximum number of random values to sample for calculation of median of
// 	// random elements
// 	nbRandoms = 1000
// )
//
// func (p stPlane) Pivot() int { return kdtree.Partition(p, kdtree.MedianOfRandoms(p, nbRandoms)) }

//Slice slicer
func (p stPlane) Slice(start, end int) kdtree.SortSlicer { p.Stations = p.Stations[start:end]; return p }

//Swap swapper
func (p stPlane) Swap(i, j int) {
	p.Stations[i], p.Stations[j] = p.Stations[j], p.Stations[i]
}
