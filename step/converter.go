package step

import (
	"fmt"

	"github.com/deadsy/sdfx/sdf"
	v3 "github.com/deadsy/sdfx/vec/v3"
)

// MeshConverter converts a triangle mesh to STEP BREP entities
type MeshConverter struct {
	entities  []Entity
	idCounter int

	// Cache for deduplication
	pointCache  map[v3.Vec]int
	edgeCache   map[edgeKey]int
	normalCache map[v3.Vec]int
}

type edgeKey struct {
	v1, v2 v3.Vec
}

func newEdgeKey(v1, v2 v3.Vec) edgeKey {
	// Normalize edge key by ordering vertices
	if v1.X < v2.X || (v1.X == v2.X && v1.Y < v2.Y) ||
		(v1.X == v2.X && v1.Y == v2.Y && v1.Z < v2.Z) {
		return edgeKey{v1, v2}
	}
	return edgeKey{v2, v1}
}

// NewMeshConverter creates a new mesh converter
func NewMeshConverter() *MeshConverter {
	return &MeshConverter{
		entities:    make([]Entity, 0),
		idCounter:   1,
		pointCache:  make(map[v3.Vec]int),
		edgeCache:   make(map[edgeKey]int),
		normalCache: make(map[v3.Vec]int),
	}
}

// addEntity adds an entity and assigns it an ID
func (c *MeshConverter) addEntity(e Entity) int {
	e.SetID(c.idCounter)
	c.entities = append(c.entities, e)
	c.idCounter++
	return e.ID()
}

// getOrCreatePoint creates or retrieves a cached CARTESIAN_POINT
func (c *MeshConverter) getOrCreatePoint(p v3.Vec) int {
	// Check cache with tolerance
	const tolerance = 1e-6
	for cached, id := range c.pointCache {
		if cached.Equals(p, tolerance) {
			return id
		}
	}

	// Create new point
	point := &CartesianPoint{
		Name:        "",
		Coordinates: []float64{p.X, p.Y, p.Z},
	}
	id := c.addEntity(point)
	c.pointCache[p] = id
	return id
}

// getOrCreateDirection creates or retrieves a cached DIRECTION
func (c *MeshConverter) getOrCreateDirection(d v3.Vec) int {
	d = d.Normalize()

	// Check cache
	if id, ok := c.normalCache[d]; ok {
		return id
	}

	// Create new direction
	dir := &Direction{
		Name:            "",
		DirectionRatios: []float64{d.X, d.Y, d.Z},
	}
	id := c.addEntity(dir)
	c.normalCache[d] = id
	return id
}

// createAxis2Placement creates an AXIS2_PLACEMENT_3D at origin
func (c *MeshConverter) createAxis2Placement(origin v3.Vec, zAxis, xAxis v3.Vec) int {
	locID := c.getOrCreatePoint(origin)
	axisID := c.getOrCreateDirection(zAxis)
	refDirID := c.getOrCreateDirection(xAxis)

	placement := &Axis2Placement3D{
		Name:         "",
		Location:     locID,
		Axis:         axisID,
		RefDirection: refDirID,
	}
	return c.addEntity(placement)
}

// createVertexPoint creates a VERTEX_POINT
func (c *MeshConverter) createVertexPoint(p v3.Vec) int {
	pointID := c.getOrCreatePoint(p)
	vertex := &VertexPoint{
		Name:           "",
		VertexGeometry: pointID,
	}
	return c.addEntity(vertex)
}

// createEdgeCurve creates an EDGE_CURVE with a LINE
func (c *MeshConverter) createEdgeCurve(v1, v2 v3.Vec) int {
	// Check cache
	key := newEdgeKey(v1, v2)
	if id, ok := c.edgeCache[key]; ok {
		return id
	}

	// Create vertices
	vertex1ID := c.createVertexPoint(v1)
	vertex2ID := c.createVertexPoint(v2)

	// Create line geometry
	startPointID := c.getOrCreatePoint(v1)
	direction := v2.Sub(v1).Normalize()
	dirID := c.getOrCreateDirection(direction)
	magnitude := v2.Sub(v1).Length()

	vector := &Vector{
		Name:        "",
		Orientation: dirID,
		Magnitude:   magnitude,
	}
	vectorID := c.addEntity(vector)

	line := &Line{
		Name: "",
		Pnt:  startPointID,
		Dir:  vectorID,
	}
	lineID := c.addEntity(line)

	// Create edge curve
	edge := &EdgeCurve{
		Name:         "",
		EdgeStart:    vertex1ID,
		EdgeEnd:      vertex2ID,
		EdgeGeometry: lineID,
		SameSense:    true,
	}
	edgeID := c.addEntity(edge)

	// Cache the edge
	c.edgeCache[key] = edgeID
	return edgeID
}

// createTriangleFace creates an ADVANCED_FACE from a triangle
func (c *MeshConverter) createTriangleFace(t *sdf.Triangle3) int {
	// Get triangle vertices
	v0, v1, v2 := t[0], t[1], t[2]

	// Create edges for the triangle
	edge1ID := c.createEdgeCurve(v0, v1)
	edge2ID := c.createEdgeCurve(v1, v2)
	edge3ID := c.createEdgeCurve(v2, v0)

	// Create oriented edges
	orientedEdge1 := &OrientedEdge{
		Name:        "",
		EdgeElement: edge1ID,
		Orientation: true,
	}
	oe1ID := c.addEntity(orientedEdge1)

	orientedEdge2 := &OrientedEdge{
		Name:        "",
		EdgeElement: edge2ID,
		Orientation: true,
	}
	oe2ID := c.addEntity(orientedEdge2)

	orientedEdge3 := &OrientedEdge{
		Name:        "",
		EdgeElement: edge3ID,
		Orientation: true,
	}
	oe3ID := c.addEntity(orientedEdge3)

	// Create edge loop
	edgeLoop := &EdgeLoop{
		Name:     "",
		EdgeList: []int{oe1ID, oe2ID, oe3ID},
	}
	loopID := c.addEntity(edgeLoop)

	// Create face outer bound
	faceBound := &FaceOuterBound{
		Name:        "",
		Bound:       loopID,
		Orientation: true,
	}
	boundID := c.addEntity(faceBound)

	// Create plane surface for the triangle
	normal := t.Normal()
	origin := v0

	// Calculate reference direction (X-axis) in the plane
	edge1 := v1.Sub(v0).Normalize()
	xAxis := edge1
	zAxis := normal

	planeAxisID := c.createAxis2Placement(origin, zAxis, xAxis)

	plane := &Plane{
		Name:     "",
		Position: planeAxisID,
	}
	planeID := c.addEntity(plane)

	// Create advanced face
	face := &AdvancedFace{
		Name:         "",
		Bounds:       []int{boundID},
		FaceGeometry: planeID,
		SameSense:    true,
	}
	return c.addEntity(face)
}

// ConvertMesh converts a triangle mesh to STEP entities
func (c *MeshConverter) ConvertMesh(mesh []*sdf.Triangle3, name string) []Entity {
	fmt.Printf("ConvertMesh: Starting conversion of %d triangles\n", len(mesh))

	// Reset for new conversion
	c.entities = make([]Entity, 0)
	c.idCounter = 1
	c.pointCache = make(map[v3.Vec]int)
	c.edgeCache = make(map[edgeKey]int)
	c.normalCache = make(map[v3.Vec]int)

	fmt.Println("ConvertMesh: Creating application context...")
	// Create application context
	appContext := &ApplicationContext{
		Application: "sdfx STEP Writer",
	}
	appContextID := c.addEntity(appContext)

	// Create units
	lengthUnit := &LengthUnit{}
	lengthUnitID := c.addEntity(lengthUnit)

	planeAngleUnit := &PlaneAngleUnit{}
	planeAngleUnitID := c.addEntity(planeAngleUnit)

	solidAngleUnit := &SolidAngleUnit{}
	solidAngleUnitID := c.addEntity(solidAngleUnit)

	// Create uncertainty
	uncertainty := &UncertaintyMeasureWithUnit{
		Value:       1e-6,
		Unit:        lengthUnitID,
		Name:        "DISTANCE_ACCURACY_VALUE",
		Description: "Maximum model space distance between geometric entities",
	}
	uncertaintyID := c.addEntity(uncertainty)

	// Create geometric representation context
	geomContext := &GeometricRepresentationContext{
		ContextIdentifier:        "",
		ContextType:              "3D",
		CoordinateSpaceDimension: 3,
		Uncertainty:              []int{uncertaintyID},
		Units:                    []int{lengthUnitID, planeAngleUnitID, solidAngleUnitID},
	}
	geomContextID := c.addEntity(geomContext)

	// Create product hierarchy
	productContext := &ProductContext{
		Name:             "",
		FrameOfReference: appContextID,
		DisciplineType:   "mechanical",
	}
	productContextID := c.addEntity(productContext)

	product := &Product{
		Name:             name,
		Description:      "Generated from sdfx",
		FrameOfReference: []int{productContextID},
	}
	productID := c.addEntity(product)

	productDefFormation := &ProductDefinitionFormation{
		Description: "",
		OfProduct:   productID,
	}
	pdfID := c.addEntity(productDefFormation)

	productDefContext := &ProductDefinitionContext{
		Name:             "",
		FrameOfReference: appContextID,
		LifeCycleStage:   "design",
	}
	pdcID := c.addEntity(productDefContext)

	productDef := &ProductDefinition{
		Description:      "",
		Formation:        pdfID,
		FrameOfReference: pdcID,
	}
	pdID := c.addEntity(productDef)

	productDefShape := &ProductDefinitionShape{
		Name:        "",
		Description: "",
		Definition:  pdID,
	}
	pdsID := c.addEntity(productDefShape)

	// Convert triangles to faces
	fmt.Printf("ConvertMesh: Converting %d triangles to faces...\n", len(mesh))
	faceIDs := make([]int, 0, len(mesh))
	for i, triangle := range mesh {
		if i%100 == 0 {
			fmt.Printf("ConvertMesh: Processing triangle %d/%d\n", i, len(mesh))
		}
		if !triangle.Degenerate(1e-9) {
			faceID := c.createTriangleFace(triangle)
			faceIDs = append(faceIDs, faceID)
		}
	}
	fmt.Printf("ConvertMesh: Created %d faces\n", len(faceIDs))

	fmt.Println("ConvertMesh: Creating final entities...")
	// Create closed shell
	closedShell := &ClosedShell{
		Name:  "",
		Faces: faceIDs,
	}
	shellID := c.addEntity(closedShell)
	fmt.Printf("ConvertMesh: Created closed shell with ID %d\n", shellID)

	// Create manifold solid brep
	solidBrep := &ManifoldSolidBrep{
		Name:  "",
		Outer: shellID,
	}
	brepID := c.addEntity(solidBrep)
	fmt.Printf("ConvertMesh: Created solid BREP with ID %d\n", brepID)

	// Create placement for the solid
	origin := v3.Vec{X: 0, Y: 0, Z: 0}
	zAxis := v3.Vec{X: 0, Y: 0, Z: 1}
	xAxis := v3.Vec{X: 1, Y: 0, Z: 0}

	placement := &Axis2Placement3D{
		Name:         "",
		Location:     c.getOrCreatePoint(origin),
		Axis:         c.getOrCreateDirection(zAxis),
		RefDirection: c.getOrCreateDirection(xAxis),
	}
	mainPlacementID := c.addEntity(placement)
	fmt.Printf("ConvertMesh: Created placement with ID %d\n", mainPlacementID)

	// Create advanced brep shape representation
	advBrep := &AdvancedBrepShapeRepresentation{
		Name:           "",
		Items:          []int{brepID, mainPlacementID},
		ContextOfItems: geomContextID,
	}
	advBrepID := c.addEntity(advBrep)
	fmt.Printf("ConvertMesh: Created advanced BREP with ID %d\n", advBrepID)

	// Create shape definition representation
	shapeDefRep := &ShapeDefinitionRepresentation{
		Definition:         pdsID,
		UsedRepresentation: advBrepID,
	}
	c.addEntity(shapeDefRep)

	fmt.Printf("ConvertMesh: Conversion complete with %d entities\n", len(c.entities))
	return c.entities
}

// OptimizeMesh performs mesh optimization before conversion
func OptimizeMesh(mesh []*sdf.Triangle3) []*sdf.Triangle3 {
	// Remove degenerate triangles
	optimized := make([]*sdf.Triangle3, 0, len(mesh))
	for _, t := range mesh {
		if !t.Degenerate(1e-9) {
			optimized = append(optimized, t)
		}
	}

	// Additional optimizations could include:
	// - Vertex welding
	// - Face merging for coplanar triangles
	// - Edge collapse for small edges

	return optimized
}
