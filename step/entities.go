// Package step implements STEP AP214 file generation for sdfx
package step

import (
	"fmt"
	"strings"
)

// Entity represents a STEP entity with an ID
type Entity interface {
	ID() int
	SetID(int)
	String() string
}

// BaseEntity provides common entity functionality
type BaseEntity struct {
	id int
}

func (e *BaseEntity) ID() int      { return e.id }
func (e *BaseEntity) SetID(id int) { e.id = id }

// ApplicationContext represents APPLICATION_CONTEXT entity
type ApplicationContext struct {
	BaseEntity
	Application string
}

func (e *ApplicationContext) String() string {
	return fmt.Sprintf("#%d=APPLICATION_CONTEXT('%s');", e.id, e.Application)
}

// Product represents PRODUCT entity
type Product struct {
	BaseEntity
	Name             string
	Description      string
	FrameOfReference []int // refs to PRODUCT_CONTEXT
}

func (e *Product) String() string {
	refs := formatRefs(e.FrameOfReference)
	return fmt.Sprintf("#%d=PRODUCT('','%s','%s',(%s));", e.id, e.Name, e.Description, refs)
}

// ProductContext represents PRODUCT_CONTEXT entity
type ProductContext struct {
	BaseEntity
	Name             string
	FrameOfReference int // ref to APPLICATION_CONTEXT
	DisciplineType   string
}

func (e *ProductContext) String() string {
	return fmt.Sprintf("#%d=PRODUCT_CONTEXT('%s',#%d,'%s');",
		e.id, e.Name, e.FrameOfReference, e.DisciplineType)
}

// ProductDefinitionFormation represents PRODUCT_DEFINITION_FORMATION entity
type ProductDefinitionFormation struct {
	BaseEntity
	Description string
	OfProduct   int // ref to PRODUCT
}

func (e *ProductDefinitionFormation) String() string {
	return fmt.Sprintf("#%d=PRODUCT_DEFINITION_FORMATION('','%s',#%d);",
		e.id, e.Description, e.OfProduct)
}

// ProductDefinitionContext represents PRODUCT_DEFINITION_CONTEXT entity
type ProductDefinitionContext struct {
	BaseEntity
	Name             string
	FrameOfReference int // ref to APPLICATION_CONTEXT
	LifeCycleStage   string
}

func (e *ProductDefinitionContext) String() string {
	return fmt.Sprintf("#%d=PRODUCT_DEFINITION_CONTEXT('%s',#%d,'%s');",
		e.id, e.Name, e.FrameOfReference, e.LifeCycleStage)
}

// ProductDefinition represents PRODUCT_DEFINITION entity
type ProductDefinition struct {
	BaseEntity
	Description      string
	Formation        int // ref to PRODUCT_DEFINITION_FORMATION
	FrameOfReference int // ref to PRODUCT_DEFINITION_CONTEXT
}

func (e *ProductDefinition) String() string {
	return fmt.Sprintf("#%d=PRODUCT_DEFINITION('','%s',#%d,#%d);",
		e.id, e.Description, e.Formation, e.FrameOfReference)
}

// ProductDefinitionShape represents PRODUCT_DEFINITION_SHAPE entity
type ProductDefinitionShape struct {
	BaseEntity
	Name        string
	Description string
	Definition  int // ref to PRODUCT_DEFINITION
}

func (e *ProductDefinitionShape) String() string {
	return fmt.Sprintf("#%d=PRODUCT_DEFINITION_SHAPE('%s','%s',#%d);",
		e.id, e.Name, e.Description, e.Definition)
}

// ShapeDefinitionRepresentation represents SHAPE_DEFINITION_REPRESENTATION entity
type ShapeDefinitionRepresentation struct {
	BaseEntity
	Definition         int // ref to PRODUCT_DEFINITION_SHAPE
	UsedRepresentation int // ref to ADVANCED_BREP_SHAPE_REPRESENTATION
}

func (e *ShapeDefinitionRepresentation) String() string {
	return fmt.Sprintf("#%d=SHAPE_DEFINITION_REPRESENTATION(#%d,#%d);",
		e.id, e.Definition, e.UsedRepresentation)
}

// AdvancedBrepShapeRepresentation represents ADVANCED_BREP_SHAPE_REPRESENTATION entity
type AdvancedBrepShapeRepresentation struct {
	BaseEntity
	Name           string
	Items          []int // refs to REPRESENTATION_ITEM
	ContextOfItems int   // ref to GEOMETRIC_REPRESENTATION_CONTEXT
}

func (e *AdvancedBrepShapeRepresentation) String() string {
	items := formatRefs(e.Items)
	return fmt.Sprintf("#%d=ADVANCED_BREP_SHAPE_REPRESENTATION('%s',(%s),#%d);",
		e.id, e.Name, items, e.ContextOfItems)
}

// ManifoldSolidBrep represents MANIFOLD_SOLID_BREP entity
type ManifoldSolidBrep struct {
	BaseEntity
	Name  string
	Outer int // ref to CLOSED_SHELL
}

func (e *ManifoldSolidBrep) String() string {
	return fmt.Sprintf("#%d=MANIFOLD_SOLID_BREP('%s',#%d);", e.id, e.Name, e.Outer)
}

// ClosedShell represents CLOSED_SHELL entity
type ClosedShell struct {
	BaseEntity
	Name  string
	Faces []int // refs to ADVANCED_FACE
}

func (e *ClosedShell) String() string {
	faces := formatRefs(e.Faces)
	return fmt.Sprintf("#%d=CLOSED_SHELL('%s',(%s));", e.id, e.Name, faces)
}

// AdvancedFace represents ADVANCED_FACE entity
type AdvancedFace struct {
	BaseEntity
	Name         string
	Bounds       []int // refs to FACE_BOUND/FACE_OUTER_BOUND
	FaceGeometry int   // ref to SURFACE
	SameSense    bool
}

func (e *AdvancedFace) String() string {
	bounds := formatRefs(e.Bounds)
	sense := formatBool(e.SameSense)
	return fmt.Sprintf("#%d=ADVANCED_FACE('%s',(%s),#%d,%s);",
		e.id, e.Name, bounds, e.FaceGeometry, sense)
}

// FaceOuterBound represents FACE_OUTER_BOUND entity
type FaceOuterBound struct {
	BaseEntity
	Name        string
	Bound       int // ref to EDGE_LOOP
	Orientation bool
}

func (e *FaceOuterBound) String() string {
	orient := formatBool(e.Orientation)
	return fmt.Sprintf("#%d=FACE_OUTER_BOUND('%s',#%d,%s);", e.id, e.Name, e.Bound, orient)
}

// FaceBound represents FACE_BOUND entity
type FaceBound struct {
	BaseEntity
	Name        string
	Bound       int // ref to EDGE_LOOP
	Orientation bool
}

func (e *FaceBound) String() string {
	orient := formatBool(e.Orientation)
	return fmt.Sprintf("#%d=FACE_BOUND('%s',#%d,%s);", e.id, e.Name, e.Bound, orient)
}

// EdgeLoop represents EDGE_LOOP entity
type EdgeLoop struct {
	BaseEntity
	Name     string
	EdgeList []int // refs to ORIENTED_EDGE
}

func (e *EdgeLoop) String() string {
	edges := formatRefs(e.EdgeList)
	return fmt.Sprintf("#%d=EDGE_LOOP('%s',(%s));", e.id, e.Name, edges)
}

// OrientedEdge represents ORIENTED_EDGE entity
type OrientedEdge struct {
	BaseEntity
	Name        string
	EdgeElement int // ref to EDGE_CURVE
	Orientation bool
}

func (e *OrientedEdge) String() string {
	orient := formatBool(e.Orientation)
	return fmt.Sprintf("#%d=ORIENTED_EDGE('%s',*,*,#%d,%s);",
		e.id, e.Name, e.EdgeElement, orient)
}

// EdgeCurve represents EDGE_CURVE entity
type EdgeCurve struct {
	BaseEntity
	Name         string
	EdgeStart    int // ref to VERTEX_POINT
	EdgeEnd      int // ref to VERTEX_POINT
	EdgeGeometry int // ref to CURVE
	SameSense    bool
}

func (e *EdgeCurve) String() string {
	sense := formatBool(e.SameSense)
	return fmt.Sprintf("#%d=EDGE_CURVE('%s',#%d,#%d,#%d,%s);",
		e.id, e.Name, e.EdgeStart, e.EdgeEnd, e.EdgeGeometry, sense)
}

// VertexPoint represents VERTEX_POINT entity
type VertexPoint struct {
	BaseEntity
	Name           string
	VertexGeometry int // ref to CARTESIAN_POINT
}

func (e *VertexPoint) String() string {
	return fmt.Sprintf("#%d=VERTEX_POINT('%s',#%d);", e.id, e.Name, e.VertexGeometry)
}

// CartesianPoint represents CARTESIAN_POINT entity
type CartesianPoint struct {
	BaseEntity
	Name        string
	Coordinates []float64
}

func (e *CartesianPoint) String() string {
	coords := formatFloats(e.Coordinates)
	return fmt.Sprintf("#%d=CARTESIAN_POINT('%s',(%s));", e.id, e.Name, coords)
}

// Direction represents DIRECTION entity
type Direction struct {
	BaseEntity
	Name            string
	DirectionRatios []float64
}

func (e *Direction) String() string {
	ratios := formatFloats(e.DirectionRatios)
	return fmt.Sprintf("#%d=DIRECTION('%s',(%s));", e.id, e.Name, ratios)
}

// Vector represents VECTOR entity
type Vector struct {
	BaseEntity
	Name        string
	Orientation int // ref to DIRECTION
	Magnitude   float64
}

func (e *Vector) String() string {
	return fmt.Sprintf("#%d=VECTOR('%s',#%d,%.6f);", e.id, e.Name, e.Orientation, e.Magnitude)
}

// Axis2Placement3D represents AXIS2_PLACEMENT_3D entity
type Axis2Placement3D struct {
	BaseEntity
	Name         string
	Location     int // ref to CARTESIAN_POINT
	Axis         int // ref to DIRECTION
	RefDirection int // ref to DIRECTION
}

func (e *Axis2Placement3D) String() string {
	return fmt.Sprintf("#%d=AXIS2_PLACEMENT_3D('%s',#%d,#%d,#%d);",
		e.id, e.Name, e.Location, e.Axis, e.RefDirection)
}

// Line represents LINE entity
type Line struct {
	BaseEntity
	Name string
	Pnt  int // ref to CARTESIAN_POINT
	Dir  int // ref to VECTOR
}

func (e *Line) String() string {
	return fmt.Sprintf("#%d=LINE('%s',#%d,#%d);", e.id, e.Name, e.Pnt, e.Dir)
}

// Circle represents CIRCLE entity
type Circle struct {
	BaseEntity
	Name     string
	Position int // ref to AXIS2_PLACEMENT_3D
	Radius   float64
}

func (e *Circle) String() string {
	return fmt.Sprintf("#%d=CIRCLE('%s',#%d,%.6f);", e.id, e.Name, e.Position, e.Radius)
}

// Plane represents PLANE entity
type Plane struct {
	BaseEntity
	Name     string
	Position int // ref to AXIS2_PLACEMENT_3D
}

func (e *Plane) String() string {
	return fmt.Sprintf("#%d=PLANE('%s',#%d);", e.id, e.Name, e.Position)
}

// CylindricalSurface represents CYLINDRICAL_SURFACE entity
type CylindricalSurface struct {
	BaseEntity
	Name     string
	Position int // ref to AXIS2_PLACEMENT_3D
	Radius   float64
}

func (e *CylindricalSurface) String() string {
	return fmt.Sprintf("#%d=CYLINDRICAL_SURFACE('%s',#%d,%.6f);",
		e.id, e.Name, e.Position, e.Radius)
}

// ConicalSurface represents CONICAL_SURFACE entity
type ConicalSurface struct {
	BaseEntity
	Name      string
	Position  int // ref to AXIS2_PLACEMENT_3D
	Radius    float64
	SemiAngle float64
}

func (e *ConicalSurface) String() string {
	return fmt.Sprintf("#%d=CONICAL_SURFACE('%s',#%d,%.6f,%.6f);",
		e.id, e.Name, e.Position, e.Radius, e.SemiAngle)
}

// SphericalSurface represents SPHERICAL_SURFACE entity
type SphericalSurface struct {
	BaseEntity
	Name     string
	Position int // ref to AXIS2_PLACEMENT_3D
	Radius   float64
}

func (e *SphericalSurface) String() string {
	return fmt.Sprintf("#%d=SPHERICAL_SURFACE('%s',#%d,%.6f);",
		e.id, e.Name, e.Position, e.Radius)
}

// ToroidalSurface represents TOROIDAL_SURFACE entity
type ToroidalSurface struct {
	BaseEntity
	Name        string
	Position    int // ref to AXIS2_PLACEMENT_3D
	MajorRadius float64
	MinorRadius float64
}

func (e *ToroidalSurface) String() string {
	return fmt.Sprintf("#%d=TOROIDAL_SURFACE('%s',#%d,%.6f,%.6f);",
		e.id, e.Name, e.Position, e.MajorRadius, e.MinorRadius)
}

// BSplineCurveWithKnots represents B_SPLINE_CURVE_WITH_KNOTS entity
type BSplineCurveWithKnots struct {
	BaseEntity
	Name               string
	Degree             int
	ControlPointsList  []int // refs to CARTESIAN_POINT
	CurveForm          string
	ClosedCurve        bool
	SelfIntersect      bool
	KnotMultiplicities []int
	Knots              []float64
	KnotSpec           string
}

func (e *BSplineCurveWithKnots) String() string {
	points := formatRefs(e.ControlPointsList)
	closed := formatLogical(e.ClosedCurve)
	selfInt := formatLogical(e.SelfIntersect)
	mults := formatInts(e.KnotMultiplicities)
	knots := formatFloats(e.Knots)

	return fmt.Sprintf("#%d=B_SPLINE_CURVE_WITH_KNOTS('%s',%d,(%s),%s,%s,%s,(%s),(%s),%s);",
		e.id, e.Name, e.Degree, points, e.CurveForm, closed, selfInt, mults, knots, e.KnotSpec)
}

// Complex entity types

// GeometricRepresentationContext represents GEOMETRIC_REPRESENTATION_CONTEXT entity
type GeometricRepresentationContext struct {
	BaseEntity
	ContextIdentifier        string
	ContextType              string
	CoordinateSpaceDimension int
	Uncertainty              []int // refs to UNCERTAINTY_MEASURE_WITH_UNIT
	Units                    []int // refs to UNIT entities
}

func (e *GeometricRepresentationContext) String() string {
	// Complex entity combining multiple contexts
	uncertainty := formatRefs(e.Uncertainty)
	units := formatRefs(e.Units)

	parts := []string{
		fmt.Sprintf("GEOMETRIC_REPRESENTATION_CONTEXT(%d)", e.CoordinateSpaceDimension),
		fmt.Sprintf("GLOBAL_UNCERTAINTY_ASSIGNED_CONTEXT((%s))", uncertainty),
		fmt.Sprintf("GLOBAL_UNIT_ASSIGNED_CONTEXT((%s))", units),
		fmt.Sprintf("REPRESENTATION_CONTEXT('%s','%s')", e.ContextIdentifier, e.ContextType),
	}

	return fmt.Sprintf("#%d=(%s);", e.id, strings.Join(parts, "\n"))
}

// UncertaintyMeasureWithUnit represents UNCERTAINTY_MEASURE_WITH_UNIT entity
type UncertaintyMeasureWithUnit struct {
	BaseEntity
	Value       float64
	Unit        int // ref to UNIT
	Name        string
	Description string
}

func (e *UncertaintyMeasureWithUnit) String() string {
	return fmt.Sprintf("#%d=UNCERTAINTY_MEASURE_WITH_UNIT(LENGTH_MEASURE(%.6E),#%d,'%s','%s');",
		e.id, e.Value, e.Unit, e.Name, e.Description)
}

// LengthUnit represents LENGTH_UNIT complex entity
type LengthUnit struct {
	BaseEntity
}

func (e *LengthUnit) String() string {
	return fmt.Sprintf("#%d=(LENGTH_UNIT()\nNAMED_UNIT(*)\nSI_UNIT(.MILLI.,.METRE.));", e.id)
}

// PlaneAngleUnit represents PLANE_ANGLE_UNIT complex entity
type PlaneAngleUnit struct {
	BaseEntity
}

func (e *PlaneAngleUnit) String() string {
	return fmt.Sprintf("#%d=(NAMED_UNIT(*)\nPLANE_ANGLE_UNIT()\nSI_UNIT($,.RADIAN.));", e.id)
}

// SolidAngleUnit represents SOLID_ANGLE_UNIT complex entity
type SolidAngleUnit struct {
	BaseEntity
}

func (e *SolidAngleUnit) String() string {
	return fmt.Sprintf("#%d=(NAMED_UNIT(*)\nSI_UNIT($,.STERADIAN.)\nSOLID_ANGLE_UNIT());", e.id)
}

// Helper functions

func formatRefs(refs []int) string {
	strs := make([]string, len(refs))
	for i, ref := range refs {
		strs[i] = fmt.Sprintf("#%d", ref)
	}
	return strings.Join(strs, ",")
}

func formatFloats(vals []float64) string {
	strs := make([]string, len(vals))
	for i, val := range vals {
		strs[i] = fmt.Sprintf("%.6f", val)
	}
	return strings.Join(strs, ",")
}

func formatInts(vals []int) string {
	strs := make([]string, len(vals))
	for i, val := range vals {
		strs[i] = fmt.Sprintf("%d", val)
	}
	return strings.Join(strs, ",")
}

func formatBool(b bool) string {
	if b {
		return ".T."
	}
	return ".F."
}

func formatLogical(b bool) string {
	if b {
		return ".T."
	}
	return ".F."
}
