# Generic comparison

Implementation: `internal/notation/compare.go`. It requires an independently
frozen scale, emits G/T/S/L/D separately, and has no total. CLI comparison is
doubly locked by an explicit flag and the repository authorization file.
`compare-classes` is present but cannot run while authorization is false.
