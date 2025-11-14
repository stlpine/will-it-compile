package docker

import (
	"archive/tar"
	"bytes"
	"io"
	"time"
)

// createSourceTar creates a tar archive containing the source code.
func createSourceTar(sourceCode, filename string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Create tar header
	header := &tar.Header{
		Name:    filename,
		Mode:    0o644,
		Size:    int64(len(sourceCode)),
		ModTime: time.Now(),
	}

	// Write header
	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}

	// Write content
	if _, err := tw.Write([]byte(sourceCode)); err != nil {
		return nil, err
	}

	// Close tar writer
	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}
