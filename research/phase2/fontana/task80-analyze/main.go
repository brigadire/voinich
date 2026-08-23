// task80-analyze validates the frozen provenance-aware algebra and reports
// checksum failures instead of regenerating or mutating frozen artifacts.
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"zcore.dev/voinich/internal/fontanaalgebra"
)

func main() {
	root := filepath.Join("research", "phase2", "fontana", "task80")
	algebra, err := fontanaalgebra.LoadAlgebra(filepath.Join(root, "FONTANA_OPERATION_ALGEBRA_V1.json"))
	if err != nil {
		log.Fatal(err)
	}
	manifest, err := fontanaalgebra.LoadFrozenManifest(filepath.Join(root, "FONTANA_MODELS_FROZEN_V1.json"))
	if err != nil {
		log.Fatal(err)
	}
	if err := fontanaalgebra.VerifyChecksums(".", manifest); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("validated %s: %d types, %d operations, %d compositions, %d frozen models\n",
		algebra.Version, len(algebra.DomainTypes), len(algebra.Operations), len(algebra.Compositions), len(manifest.Models))
}
