//-----------------------------------------------------------------------------
/*

Axoloti Board Mounting Kit

*/
//-----------------------------------------------------------------------------

package main

import (
	"log"

	"github.com/deadsy/sdfx/obj"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/conv"
	v2 "github.com/deadsy/sdfx/vec/v2"
	v3 "github.com/deadsy/sdfx/vec/v3"
)

//-----------------------------------------------------------------------------

// material shrinkage
var shrink = 1.0 / 0.999 // PLA ~0.1%
//var shrink = 1.0/0.995; // ABS ~0.5%

//-----------------------------------------------------------------------------

const frontPanelThickness = 3.0
const frontPanelLength = 170.0
const frontPanelHeight = 50.0
const frontPanelYOffset = 15.0

const baseWidth = 50.0
const baseLength = 170.0
const baseThickness = 3.0

const baseFootWidth = 10.0
const baseFootCornerRadius = 3.0

const pcbWidth = 50.0
const pcbLength = 160.0

const pillarHeight = 16.8

//-----------------------------------------------------------------------------

// multiple standoffs
func standoffs() (sdf.SDF3, error) {

	k := &obj.StandoffParms{
		PillarHeight:   pillarHeight,
		PillarDiameter: 6.0,
		HoleDepth:      10.0,
		HoleDiameter:   2.4,
	}

	zOfs := 0.5 * (pillarHeight + baseThickness)

	// from the board mechanicals
	positions := v3.VecSet{
		v3.Vec{X: 3.5, Y: 10.0, Z: zOfs},   // H1
		v3.Vec{X: 3.5, Y: 40.0, Z: zOfs},   // H2
		v3.Vec{X: 54.0, Y: 40.0, Z: zOfs},  // H3
		v3.Vec{X: 156.5, Y: 10.0, Z: zOfs}, // H4
		//v3.Vec{X: 54.0, Y: 10.0, Z: zOfs},  // H5
		v3.Vec{X: 156.5, Y: 40.0, Z: zOfs}, // H6
		v3.Vec{X: 44.0, Y: 10.0, Z: zOfs},  // H7
		v3.Vec{X: 116.0, Y: 10.0, Z: zOfs}, // H8
	}

	s, err := obj.Standoff3D(k)
	if err != nil {
		return nil, err
	}
	return sdf.Multi3D(s, positions), nil
}

//-----------------------------------------------------------------------------

// base returns the base mount.
func base() (sdf.SDF3, error) {
	// base
	pp := &obj.PanelParms{
		Size:         v2.Vec{X: baseLength, Y: baseWidth},
		CornerRadius: 5.0,
		HoleDiameter: 3.5,
		HoleMargin:   [4]float64{7.0, 20.0, 7.0, 20.0},
		HolePattern:  [4]string{"xx", "x", "xx", "x"},
	}
	s0, err := obj.Panel2D(pp)
	if err != nil {
		return nil, err
	}

	// cutout
	l := baseLength - (2.0 * baseFootWidth)
	w := 18.0
	s1 := sdf.Box2D(v2.Vec{X: l, Y: w}, baseFootCornerRadius)
	yOfs := 0.5 * (baseWidth - pcbWidth)
	s1 = sdf.Transform2D(s1, sdf.Translate2d(v2.Vec{X: 0, Y: yOfs}))

	s2 := sdf.Extrude3D(sdf.Difference2D(s0, s1), baseThickness)
	xOfs := 0.5 * pcbLength
	yOfs = pcbWidth - (0.5 * baseWidth)
	s2 = sdf.Transform3D(s2, sdf.Translate3d(v3.Vec{X: xOfs, Y: yOfs, Z: 0}))

	// standoffs
	s3, err := standoffs()
	if err != nil {
		return nil, err
	}

	s4 := sdf.Union3D(s2, s3)
	s4.(*sdf.UnionSDF3).SetMin(sdf.PolyMin(3.0))

	return s4, nil
}

//-----------------------------------------------------------------------------
// front panel cutouts

type panelHole struct {
	center v2.Vec   // center of hole
	hole   sdf.SDF2 // 2d hole
}

// button positions
const pbX = 53.0

var pb0 = v2.Vec{X: pbX, Y: 0.8}
var pb1 = v2.Vec{X: pbX + 5.334, Y: 0.8}

// panelCutouts returns the 2D front panel cutouts
func panelCutouts() (sdf.SDF2, error) {

	sMidi, err := sdf.Circle2D(0.5 * 17.0)
	if err != nil {
		return nil, err
	}
	sJack0, err := sdf.Circle2D(0.5 * 11.5)
	if err != nil {
		return nil, err
	}
	sJack1, err := sdf.Circle2D(0.5 * 5.5)
	if err != nil {
		return nil, err
	}

	sLed := sdf.Box2D(v2.Vec{X: 1.6, Y: 1.6}, 0)

	k := obj.FingerButtonParms{
		Width:  4.0,
		Gap:    0.6,
		Length: 20.0,
	}
	fb, err := obj.FingerButton2D(&k)
	if err != nil {
		return nil, err
	}
	sButton := sdf.Transform2D(fb, sdf.Rotate2d(sdf.DtoR(-90)))

	jackX := 123.0
	midiX := 18.8
	ledX := 62.9

	holes := []panelHole{
		{center: v2.Vec{X: midiX, Y: 10.2}, hole: sMidi},                               // MIDI DIN Jack
		{center: v2.Vec{X: midiX + 20.32, Y: 10.2}, hole: sMidi},                       // MIDI DIN Jack
		{center: v2.Vec{X: jackX, Y: 8.14}, hole: sJack0},                              // 1/4" Stereo Jack
		{center: v2.Vec{X: jackX + 19.5, Y: 8.14}, hole: sJack0},                       // 1/4" Stereo Jack
		{center: v2.Vec{X: 107.6, Y: 2.3}, hole: sJack1},                               // 3.5 mm Headphone Jack
		{center: v2.Vec{X: ledX, Y: 0.5}, hole: sLed},                                  // LED
		{center: v2.Vec{X: ledX + 3.635, Y: 0.5}, hole: sLed},                          // LED
		{center: pb0, hole: sButton},                                                   // Push Button
		{center: pb1, hole: sButton},                                                   // Push Button
		{center: v2.Vec{X: 84.1, Y: 1.0}, hole: sdf.Box2D(v2.Vec{X: 16.0, Y: 7.5}, 0)}, // micro SD card
		{center: v2.Vec{X: 96.7, Y: 1.0}, hole: sdf.Box2D(v2.Vec{X: 11.0, Y: 7.5}, 0)}, // micro USB connector
		{center: v2.Vec{X: 73.1, Y: 7.1}, hole: sdf.Box2D(v2.Vec{X: 7.5, Y: 15.0}, 0)}, // fullsize USB connector
	}

	s := make([]sdf.SDF2, len(holes))
	for i, k := range holes {
		s[i] = sdf.Transform2D(k.hole, sdf.Translate2d(k.center))
	}

	return sdf.Union2D(s...), nil
}

//-----------------------------------------------------------------------------

// frontPanel returns the front panel mount.
func frontPanel() (sdf.SDF3, error) {

	// overall panel
	pp := &obj.PanelParms{
		Size:         v2.Vec{X: frontPanelLength, Y: frontPanelHeight},
		CornerRadius: 5.0,
		HoleDiameter: 3.5,
		HoleMargin:   [4]float64{5.0, 5.0, 5.0, 5.0},
		HolePattern:  [4]string{"xx", "x", "xx", "x"},
	}
	panel, err := obj.Panel2D(pp)
	if err != nil {
		return nil, err
	}

	xOfs := 0.5 * pcbLength
	yOfs := (0.5 * frontPanelHeight) - frontPanelYOffset
	panel = sdf.Transform2D(panel, sdf.Translate2d(v2.Vec{X: xOfs, Y: yOfs}))

	// extrude to 3d
	panelCutouts, err := panelCutouts()
	if err != nil {
		return nil, err
	}
	fp := sdf.Extrude3D(sdf.Difference2D(panel, panelCutouts), frontPanelThickness)

	// Add buttons to the finger button
	bHeight := 4.0
	b, _ := sdf.Cylinder3D(bHeight, 1.4, 0)
	b0 := sdf.Transform3D(b, sdf.Translate3d(conv.V2ToV3(pb0, -0.5*bHeight)))
	b1 := sdf.Transform3D(b, sdf.Translate3d(conv.V2ToV3(pb1, -0.5*bHeight)))

	return sdf.Union3D(fp, b0, b1), nil
}

//-----------------------------------------------------------------------------

func main() {

	// front panel
	s0, err := frontPanel()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	sx := sdf.Transform3D(s0, sdf.RotateY(sdf.DtoR(180.0)))
	render.ToSTL(sdf.ScaleUniform3D(sx, shrink), "panel.stl", render.NewMarchingCubesOctree(400))

	// base
	s1, err := base()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	render.ToSTL(sdf.ScaleUniform3D(s1, shrink), "base.stl", render.NewMarchingCubesOctree(400))

	// both together
	s0 = sdf.Transform3D(s0, sdf.Translate3d(v3.Vec{X: 0, Y: 80, Z: 0}))
	s3 := sdf.Union3D(s0, s1)
	render.ToSTL(sdf.ScaleUniform3D(s3, shrink), "panel_and_base.stl", render.NewMarchingCubesOctree(400))
}

//-----------------------------------------------------------------------------
