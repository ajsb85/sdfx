// Example demonstrating STEP AP214 export from sdfx models
package main

import (
	"log"

	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	v3 "github.com/deadsy/sdfx/vec/v3"
)

func main() {
	// Create a simple test model - a box with a cylinder hole
	box, err := sdf.Box3D(v3.Vec{X: 100, Y: 80, Z: 60}, 5)
	if err != nil {
		log.Fatalf("failed to create box: %v", err)
	}

	cylinder, err := sdf.Cylinder3D(65, 25, 0)
	if err != nil {
		log.Fatalf("failed to create cylinder: %v", err)
	}

	// Create the final model by subtracting the cylinder from the box
	model := sdf.Difference3D(box, cylinder)

	// Export to STL (existing functionality)
	render.ToSTL(model, "model.stl", render.NewMarchingCubesOctree(200))

	// Export to STEP with default options
	if err := render.ToSTEP(model, "model.step", render.NewMarchingCubesOctree(200)); err != nil {
		log.Fatalf("failed to export STEP: %v", err)
	}

	// Export to STEP with custom options
	opts := render.STEPOptions{
		Author:       "John Doe",
		Organization: "ACME Corp",
		ProductName:  "TestPart_v1",
	}
	if err := render.ToSTEPWithOptions(model, "model_custom.step", render.NewMarchingCubesOctree(200), opts); err != nil {
		log.Fatalf("failed to export STEP with options: %v", err)
	}

	// Alternative: Generate triangles once and export to multiple formats
	triangles := render.ToTriangles(model, render.NewMarchingCubesOctree(200))

	// Save to STL
	if err := render.SaveSTL("model2.stl", triangles); err != nil {
		log.Fatalf("failed to save STL: %v", err)
	}

	// Save to STEP
	if err := render.SaveSTEP("model2.step", triangles); err != nil {
		log.Fatalf("failed to save STEP: %v", err)
	}

	log.Println("Export completed successfully!")
}
