// Command task82a-analyze runs the Task82a corpus-scale mnemonic scaling
// experiment: it assembles corpus-scale OBSERVABLE_DOCUMENTs from the
// frozen Task81 V1.1 mechanisms, extracts frozen Fingerprint V2 features,
// and aggregates the frozen portfolio.
package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/task82a"
)

func main() {
	root := flag.String("root", ".", "repository root")
	pilot := flag.Bool("pilot", false, "run the target-blind scale convergence pilot and print its result")
	shardIndex := flag.Int("shard-index", 0, "shard index")
	shardCount := flag.Int("shard-count", 1, "shard count")
	resume := flag.Bool("resume", false, "resume from existing raw artifacts")
	verifyOnly := flag.Bool("verify-only", false, "verify existing raw artifacts without re-running")
	genManifest := flag.Bool("gen-manifest", false, "generate TASK82A_BLIND_MANIFEST.json and TASK82A_COST_MODEL.tsv, then exit")
	freeze := flag.Bool("freeze-results", false, "close the ledger and write the results freeze marker")
	flag.Parse()

	if *pilot {
		res, err := task82a.RunScaleConvergencePilot(*root, task82a.CueCapacity, task82a.PilotCheckpoints)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(res.String())
		return
	}
	if *genManifest {
		if err := task82a.GenerateManifest(*root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *freeze {
		if err := task82a.FreezeResults(*root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	err := task82a.Execute(task82a.Options{
		Root:       *root,
		ShardIndex: *shardIndex,
		ShardCount: *shardCount,
		Resume:     *resume,
		VerifyOnly: *verifyOnly,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
