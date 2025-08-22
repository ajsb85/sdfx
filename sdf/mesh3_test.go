//-----------------------------------------------------------------------------
/*

Mesh 3D Testing and Benchmarking

*/
//-----------------------------------------------------------------------------

package sdf

import (
	"testing"

	v3 "github.com/deadsy/sdfx/vec/v3"
)

//-----------------------------------------------------------------------------

func Test_Mesh3_minDistance2(t *testing.T) {

	testSet := []struct {
		t  Triangle3
		p  []v3.Vec
		d2 []float64
	}{
		{
			Triangle3{{X: 1, Y: 2, Z: 1}, {X: -4, Y: -5, Z: 1}, {X: 17, Y: -3, Z: 1}},
			[]v3.Vec{{X: 1, Y: 2, Z: 3}, {X: -4, Y: -5, Z: 6}, {X: 17, Y: -3, Z: -2}},
			[]float64{4, 25, 9},
		},
		{
			Triangle3{{X: 10, Y: 0, Z: 10}, {X: 0, Y: 0, Z: -10}, {X: -10, Y: 0, Z: 10}},
			[]v3.Vec{{X: 0, Y: 4, Z: 0}, {X: 0, Y: 0, Z: 0}, {X: 11, Y: 4, Z: 11}, {X: 0, Y: 3, Z: -11}, {X: -11, Y: 7, Z: 11}},
			[]float64{16, 0, 18, 10, 51},
		},
		{
			Triangle3{{X: 0, Y: 0, Z: 4}, {X: 0, Y: 6, Z: -2}, {X: 0, Y: -6, Z: -2}},
			[]v3.Vec{{X: 0, Y: 0, Z: 5}, {X: 0, Y: 2, Z: 2}, {X: 4, Y: 3, Z: 3}, {X: 3, Y: -3, Z: 3}, {X: -2, Y: 0, Z: -3}},
			[]float64{1, 0, 18, 11, 5},
		},
	}

	for i, test := range testSet {
		if len(test.p) != len(test.d2) {
			t.Errorf("test %d: len(p) != len(d2)", i)
		}
		triangle := test.t
		for j := 0; j < 3; j++ {
			triangle = triangle.rotateVertex()
			ti := newTriangleInfo(&triangle)
			for k, p := range test.p {
				d2 := ti.minDistance2(p)
				if !EqualFloat64(d2, test.d2[k], tolerance) {
					t.Errorf("test %d.%d: expected %f, got %f", i, k, test.d2[k], d2)
				}
			}
		}
	}

	// sanity test with random triangles
	const boxSize = 100.0
	const d2Max = 3.0 * (boxSize * boxSize)
	b := NewBox3(v3.Vec{X: 0, Y: 0, Z: 0}, v3.Vec{X: 100, Y: 100, Z: 100})
	for i := 0; i < 10000; i++ {
		x := b.RandomTriangle()
		ti := newTriangleInfo(&x)
		p := b.Random()
		d2 := ti.minDistance2(p)
		if d2 < 0 {
			t.Errorf("test %d: expected >= 0, got %f", i, d2)
		}
		if d2 > d2Max {
			t.Errorf("test %d: expected <= %f, got %f", i, d2Max, d2)
		}
	}

}

//-----------------------------------------------------------------------------
