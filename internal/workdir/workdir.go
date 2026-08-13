// Package workdir defines the repository-wide output contract for pipeline
// applications. Generated and intermediate artifacts belong below ./workdir.
package workdir

import (
	"os"
	"path/filepath"
)

const Dir = "workdir"

// Path returns a path rooted in the pipeline work directory.
func Path(elements ...string) string {
	return filepath.Join(append([]string{Dir}, elements...)...)
}

// Ensure creates the pipeline work directory.
func Ensure() error {
	return os.MkdirAll(Dir, 0o755)
}

// EnsureParent creates the parent directory required by an output path.
func EnsureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}
