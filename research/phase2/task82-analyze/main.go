package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/task82"
)

func main() {
	root := flag.String("root", ".", "repository root")
	shardIndex := flag.Int("shard-index", 0, "zero-based shard index")
	shardCount := flag.Int("shard-count", 1, "number of deterministic shards")
	resume := flag.Bool("resume", false, "reuse valid existing raw artifacts")
	verifyOnly := flag.Bool("verify-only", false, "verify existing selected raw artifacts without executing")
	flag.Parse()
	if err := task82.Execute(task82.Options{Root: *root, ShardIndex: *shardIndex, ShardCount: *shardCount, Resume: *resume, VerifyOnly: *verifyOnly}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
