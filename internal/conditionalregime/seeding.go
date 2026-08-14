package conditionalregime

// Every permutation replicate uses its own independently seeded random
// stream, derived from (base seed, a per-method/per-purpose salt, replicate
// index) rather than one shared rng advancing across the whole loop. This
// makes each replicate's result reproducible from its index alone, so a
// permutation loop interrupted after replicate i can resume at i+1 without
// needing to replay - and therefore re-spend - the first i draws.
const seedStride = 1_000_000_007

func replicateSeed(base, salt int64, index int) int64 {
	return base*seedStride + salt*104_729 + int64(index)
}

func methodSalt(method string) int64 {
	if method == "hierarchical" {
		return 2
	}
	return 1
}
