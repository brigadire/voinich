// Package residualdiagnostic explains metadata association in the frozen
// residual clustering produced by conditional-regime-analyze.  It is a
// diagnostic only: window size, K, representation and clustering method are
// fixed by the previous experiment.
package residualdiagnostic

import "io"

type Config struct {
	ConditionalDir string
	CorpusPath     string
	MetadataPath   string
	OutputDir      string
	WindowSize     int
	K              int
	Permutations   int
	Seed           int64
	Quiet          bool
	ProgressWriter io.Writer
}

type metadata struct {
	Currier, Hand, Folio []string
}

type block struct {
	ID, Joint, Currier, Hand string
	Index, Start, End        int
}

func (b block) len() int { return b.End - b.Start }

type sparse map[string]float64

type window struct {
	Currier, Hand, Joint, Folio, Block string
	BlockIndex, Start, End, Fold       int
	PhysicalStart, PhysicalEnd         int
	Raw, Residual, Whitened            sparse
	ExistingCluster                    int
}

type foldDiagnostic struct {
	WindowSize, Fold int
	Joint            string
	TrainWindows     int
	TestWindows      int
	TrainMean        norms
	TestMean         norms
}

type norms struct{ L1, L2, Linf, MeanAbs float64 }

type representationRow struct {
	Name                                                string
	Silhouette, CurrierNMI, HandNMI, JointNMI, BlockNMI float64
	ClusterRunCount                                     int
	LargestRunFraction                                  float64
}

type results struct {
	CorpusSHA, MetadataSHA              string
	TokenCount                          int
	OriginalCurrierNMI, OriginalHandNMI float64
	Windows                             []window
	Folds                               []foldDiagnostic
	Representations                     []representationRow
	Files                               map[string][][]string
	Summary                             map[string]any
}
