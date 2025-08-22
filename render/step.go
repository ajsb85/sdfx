// Package render provides STEP file export functionality for sdfx
package render

import (
	"fmt"
	"sync"

	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/step"
)

// ToSTEP renders an SDF3 to a STEP AP214 file
func ToSTEP(
	s sdf.SDF3, // sdf3 to render
	path string, // path to output file
	r Render3, // rendering method (e.g., NewMarchingCubesOctree)
) error {
	return ToSTEPWithOptions(s, path, r, STEPOptions{})
}

// STEPOptions configures STEP export
type STEPOptions struct {
	Author       string // Author name
	Organization string // Organization name
	ProductName  string // Product name (defaults to filename)
}

// ToSTEPWithOptions renders an SDF3 to a STEP AP214 file with options
func ToSTEPWithOptions(
	s sdf.SDF3,
	path string,
	r Render3,
	opts STEPOptions,
) error {
	fmt.Printf("rendering %s (%s)\n", path, r.Info(s))

	// Set default product name if not provided
	productName := opts.ProductName
	if productName == "" {
		productName = "sdfx_model"
	}

	// write the triangles to a STEP file
	var wg sync.WaitGroup
	output, err := writeSTEP(&wg, path, opts)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}
	// run the renderer
	r.Render(s, sdf.NewTriangle3Buffer(output))
	// stop the STEP writer reading on the channel
	close(output)
	// wait for the file write to complete
	wg.Wait()

	fmt.Printf("STEP export completed: %s\n", path)
	return nil
}

// writeSTEP writes a stream of triangles to a STEP file
func writeSTEP(wg *sync.WaitGroup, path string, opts STEPOptions) (chan<- []*sdf.Triangle3, error) {
	writer, err := step.NewWriter(path)
	if err != nil {
		return nil, err
	}

	// Set author information if provided
	if opts.Author != "" || opts.Organization != "" {
		author := opts.Author
		if author == "" {
			author = "Unknown"
		}
		org := opts.Organization
		if org == "" {
			org = "Unknown"
		}
		writer.SetAuthor(author, org)
	}

	// External code writes triangles to this channel.
	// This goroutine reads the channel and writes triangles to the file.
	c := make(chan []*sdf.Triangle3, 100)

	// Collect all triangles
	triangles := make([]*sdf.Triangle3, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer writer.Close()

		// Collect triangles from the channel
		for ts := range c {
			triangles = append(triangles, ts...)
			fmt.Printf("Collected batch of %d triangles (total: %d)\n", len(ts), len(triangles))
		}

		fmt.Printf("Writing %d triangles to STEP file\n", len(triangles))

		// Set default product name if not provided
		productName := opts.ProductName
		if productName == "" {
			productName = "sdfx_model"
		}

		// Write mesh to STEP file
		if err := writer.WriteMesh(triangles, productName); err != nil {
			fmt.Printf("Error writing STEP file: %v\n", err)
			return
		}

		fmt.Println("STEP file written successfully")
	}()

	return c, nil
}

// SaveSTEP writes a pre-computed triangle mesh to a STEP file
func SaveSTEP(path string, mesh []*sdf.Triangle3) error {
	return SaveSTEPWithOptions(path, mesh, STEPOptions{})
}

// SaveSTEPWithOptions writes a pre-computed triangle mesh to a STEP file with options
func SaveSTEPWithOptions(path string, mesh []*sdf.Triangle3, opts STEPOptions) error {
	writer, err := step.NewWriter(path)
	if err != nil {
		return fmt.Errorf("failed to create STEP writer: %w", err)
	}
	defer writer.Close()

	// Set author information if provided
	if opts.Author != "" || opts.Organization != "" {
		author := opts.Author
		if author == "" {
			author = "Unknown"
		}
		org := opts.Organization
		if org == "" {
			org = "Unknown"
		}
		writer.SetAuthor(author, org)
	}

	// Set default product name if not provided
	productName := opts.ProductName
	if productName == "" {
		productName = "sdfx_model"
	}

	// Write mesh to STEP file
	if err := writer.WriteMesh(mesh, productName); err != nil {
		return fmt.Errorf("failed to write mesh: %w", err)
	}

	fmt.Printf("STEP export completed: %s\n", path)
	return nil
}

// LoadSTEP loads a STEP file and converts it to a triangle mesh
// Note: This is a placeholder for future implementation
func LoadSTEP(path string) ([]*sdf.Triangle3, error) {
	// TODO: Implement STEP file parsing and conversion to triangle mesh
	// This would require:
	// 1. Parse STEP file format
	// 2. Extract BREP geometry
	// 3. Tessellate BREP to triangles
	// 4. Return triangle mesh

	return nil, fmt.Errorf("STEP import not yet implemented")
}
