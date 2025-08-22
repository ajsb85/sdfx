package step

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/deadsy/sdfx/sdf"
)

// Writer handles STEP file generation
type Writer struct {
	file       *os.File
	writer     *bufio.Writer
	converter  *MeshConverter
	fileName   string
	authorName string
	orgName    string
}

// NewWriter creates a new STEP writer
func NewWriter(path string) (*Writer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &Writer{
		file:       file,
		writer:     bufio.NewWriter(file),
		converter:  NewMeshConverter(),
		fileName:   filepath.Base(path),
		authorName: "sdfx User",
		orgName:    "sdfx Organization",
	}, nil
}

// SetAuthor sets the author information
func (w *Writer) SetAuthor(name, org string) {
	w.authorName = name
	w.orgName = org
}

// Close closes the writer and flushes any remaining data
func (w *Writer) Close() error {
	if err := w.writer.Flush(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

// writeHeader writes the STEP file header
func (w *Writer) writeHeader() error {
	header := []string{
		"ISO-10303-21;",
		"HEADER;",
		"FILE_DESCRIPTION(('STEP AP214'),'1');",
		fmt.Sprintf("FILE_NAME('%s','%s',('%s'),('%s'),'sdfx STEP Writer','sdfx','');",
			w.fileName,
			time.Now().Format("2006-01-02T15:04:05"),
			w.authorName,
			w.orgName),
		"FILE_SCHEMA(('AUTOMOTIVE_DESIGN'));",
		"ENDSEC;",
	}

	for _, line := range header {
		if _, err := w.writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// writeData writes the DATA section with entities
func (w *Writer) writeData(entities []Entity) error {
	if _, err := w.writer.WriteString("DATA;\n"); err != nil {
		return err
	}

	for _, entity := range entities {
		str := entity.String()
		// Handle multi-line entities (complex types)
		if strings.Contains(str, "\n") {
			lines := strings.Split(str, "\n")
			for i, line := range lines {
				if i < len(lines)-1 {
					if _, err := w.writer.WriteString(line + "\n"); err != nil {
						return err
					}
				} else {
					if _, err := w.writer.WriteString(line + "\n"); err != nil {
						return err
					}
				}
			}
		} else {
			if _, err := w.writer.WriteString(str + "\n"); err != nil {
				return err
			}
		}
	}

	if _, err := w.writer.WriteString("ENDSEC;\n"); err != nil {
		return err
	}

	return nil
}

// writeFooter writes the STEP file footer
func (w *Writer) writeFooter() error {
	if _, err := w.writer.WriteString("END-ISO-10303-21;\n"); err != nil {
		return err
	}
	return nil
}

// WriteMesh writes a triangle mesh to the STEP file
func (w *Writer) WriteMesh(mesh []*sdf.Triangle3, name string) error {
	fmt.Printf("WriteMesh: Starting with %d triangles\n", len(mesh))

	// Optimize mesh
	optimizedMesh := OptimizeMesh(mesh)
	fmt.Printf("WriteMesh: Optimized to %d triangles\n", len(optimizedMesh))

	// Convert mesh to STEP entities
	fmt.Println("WriteMesh: Converting to STEP entities...")
	entities := w.converter.ConvertMesh(optimizedMesh, name)
	fmt.Printf("WriteMesh: Created %d entities\n", len(entities))

	// Write STEP file
	fmt.Println("WriteMesh: Writing header...")
	if err := w.writeHeader(); err != nil {
		return err
	}

	fmt.Println("WriteMesh: Writing data section...")
	if err := w.writeData(entities); err != nil {
		return err
	}

	fmt.Println("WriteMesh: Writing footer...")
	if err := w.writeFooter(); err != nil {
		return err
	}

	fmt.Println("WriteMesh: Flushing buffer...")
	return w.writer.Flush()
}

// StreamWriter handles streaming triangle data to STEP file
type StreamWriter struct {
	writer    *Writer
	triangles []*sdf.Triangle3
	wg        *sync.WaitGroup
	input     chan []*sdf.Triangle3
	mutex     sync.Mutex
}

// NewStreamWriter creates a new streaming STEP writer
func NewStreamWriter(path string) (*StreamWriter, chan<- []*sdf.Triangle3, error) {
	writer, err := NewWriter(path)
	if err != nil {
		return nil, nil, err
	}

	input := make(chan []*sdf.Triangle3, 100) // buffered channel

	sw := &StreamWriter{
		writer:    writer,
		triangles: make([]*sdf.Triangle3, 0),
		wg:        new(sync.WaitGroup),
		input:     input,
	}

	// Start goroutine to collect triangles
	sw.wg.Add(1)
	go sw.collect()

	return sw, input, nil
}

// collect gathers triangles from the input channel
func (sw *StreamWriter) collect() {
	defer sw.wg.Done()

	for tris := range sw.input {
		sw.mutex.Lock()
		sw.triangles = append(sw.triangles, tris...)
		sw.mutex.Unlock()
		fmt.Printf("Collected %d triangles (total: %d)\n", len(tris), len(sw.triangles))
	}
	fmt.Println("Triangle collection completed")
}

// Input returns the input channel for triangles
func (sw *StreamWriter) Input() chan<- []*sdf.Triangle3 {
	return sw.input
}

// SetAuthor sets the author information
func (sw *StreamWriter) SetAuthor(name, org string) {
	sw.writer.SetAuthor(name, org)
}

// Finalize writes the collected triangles to the STEP file
func (sw *StreamWriter) Finalize(name string) error {
	fmt.Printf("Finalizing STEP file with %d triangles\n", len(sw.triangles))

	// Close input channel and wait for collection to finish
	close(sw.input)
	sw.wg.Wait()

	// Write mesh to file
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	fmt.Printf("Writing %d triangles to STEP file\n", len(sw.triangles))

	if err := sw.writer.WriteMesh(sw.triangles, name); err != nil {
		sw.writer.Close()
		return err
	}

	fmt.Println("STEP file written successfully")
	return sw.writer.Close()
}
