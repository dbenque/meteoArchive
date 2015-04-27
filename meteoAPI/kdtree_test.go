// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meteoAPI

import (
	"fmt"
	"math"
	"testing"

	"github.com/biogo/store/kdtree"
)

var (
	_ kdtree.Interface  = nbPoints{}
	_ kdtree.Comparable = nbPoint{}
)

// Randoms is the maximum number of random values to sample for calculation of median of
// random elements.
var nbRandoms = 100

// An nbPoint represents a point in a k-d space that satisfies the Comparable interface.
type nbPoint kdtree.Point

func (p nbPoint) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(nbPoint)
	return p[d] - q[d]
}
func (p nbPoint) Dims() int { return len(p) }
func (p nbPoint) Distance(c kdtree.Comparable) float64 {
	q := c.(nbPoint)
	var sum float64
	for dim, c := range p {
		d := c - q[dim]
		sum += d * d
	}
	return sum
}

// An nbPoints is a collection of point values that satisfies the Interface.
type nbPoints []nbPoint

func (p nbPoints) Index(i int) kdtree.Comparable         { return p[i] }
func (p nbPoints) Len() int                              { return len(p) }
func (p nbPoints) Pivot(d kdtree.Dim) int                { return nbPlane{nbPoints: p, Dim: d}.Pivot() }
func (p nbPoints) Slice(start, end int) kdtree.Interface { return p[start:end] }

// An nbPlane is a wrapping type that allows a Points type be pivoted on a dimension.
type nbPlane struct {
	kdtree.Dim
	nbPoints
}

func (p nbPlane) Less(i, j int) bool                     { return p.nbPoints[i][p.Dim] < p.nbPoints[j][p.Dim] }
func (p nbPlane) Pivot() int                             { return kdtree.Partition(p, kdtree.MedianOfRandoms(p, nbRandoms)) }
func (p nbPlane) Slice(start, end int) kdtree.SortSlicer { p.nbPoints = p.nbPoints[start:end]; return p }
func (p nbPlane) Swap(i, j int) {
	p.nbPoints[i], p.nbPoints[j] = p.nbPoints[j], p.nbPoints[i]
}

// [4068.4040536166076 428.8033039669525 4884.041070980361]
// [4101.327521741679 440.69480392791496 4855.36215427801]
// [4103.456389555063 447.00994428661943 4852.985552083233]
// [4086.18184048877 461.94981027688965 4866.1444018077555]
// [4176.910493418626 568.6093369018052 4777.00148125971]

// p1:= nbPoint{4068.4040536166076,428.8033039669525,4884.041070980361}
// p2:= nbPoint{4101.327521741679,440.69480392791496,4855.36215427801}
// p3:= nbPoint{4103.456389555063,447.00994428661943,4852.985552083233}
// p4:= nbPoint{4086.18184048877,461.94981027688965,4866.1444018077555}

func TestDistanceKeeper(t *testing.T) {

	p1 := nbPoint{4103.456389555063, 447.00994428661943, 4852.985552083233}
	p2 := nbPoint{4086.18184048877, 461.94981027688965, 4866.1444018077555}
	p3 := nbPoint{4101.327521741679, 440.69480392791496, 4855.36215427801}
	p4 := nbPoint{4068.4040536166076, 428.8033039669525, 4884.041070980361}

	points := nbPoints{p1, p2, p3, p4}

	p0 := nbPoint{4176.910493418626, 568.6093369018052, 4777.00148125971}

	fmt.Println(math.Sqrt(p0.Distance(p1)))
	fmt.Println(math.Sqrt(p0.Distance(p2)))
	fmt.Println(math.Sqrt(p0.Distance(p3)))
	fmt.Println(math.Sqrt(p0.Distance(p4)))

	tree := kdtree.New(points, true)
	keeper := kdtree.NewDistKeeper(168 * 168)
	tree.NearestSet(keeper, p0)

	count := 0
	for keeper.Len() > 0 {
		v := keeper.Heap.Pop()

		if c, ok := v.(kdtree.ComparableDist); ok {
			if _, ok := c.Comparable.(nbPoint); ok {
				count++
			}
		}
	}

	if count != 3 {
		t.Fatal("Should be 4, and got ", count)
	}

}
