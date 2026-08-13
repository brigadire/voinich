package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/globalregime"
	"zcore.dev/voinich/internal/workdir"
)

type intList []int

func (v *intList) String() string {
	parts := make([]string, len(*v))
	for i, n := range *v {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}
func (v *intList) Set(s string) error {
	var out []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("invalid window size %q", part)
		}
		out = append(out, n)
	}
	*v = out
	return nil
}

func main() {
	c := globalregime.Config{}
	sizes := intList{50, 100, 200, 500, 1000}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ivtt_output_1786282555007.txt", "tokenized continuous corpus")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.Var(&sizes, "window-sizes", "comma-separated sliding window sizes")
	flag.IntVar(&c.Step, "step", 0, "fixed window step (0 uses max(1, window_size/10))")
	flag.Int64Var(&c.Seed, "seed", 160016, "deterministic clustering seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	c.WindowSizes = []int(sizes)
	if err := globalregime.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
