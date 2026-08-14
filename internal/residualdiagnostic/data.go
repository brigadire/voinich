package residualdiagnostic

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/globalregime"
)

func readCorpus(path string) ([]string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return strings.Fields(string(b)), fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func readMetadata(path string) (metadata, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return metadata{}, "", err
	}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 64<<10), 16<<20)
	if !s.Scan() {
		return metadata{}, "", fmt.Errorf("empty metadata map: %s", path)
	}
	h := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
	col := map[string]int{}
	for i, v := range h {
		col[v] = i
	}
	for _, v := range []string{"token_position", "currier", "hand", "folio"} {
		if _, ok := col[v]; !ok {
			return metadata{}, "", fmt.Errorf("metadata map missing %q", v)
		}
	}
	m := metadata{}
	expected := 0
	for s.Scan() {
		r := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
		pos, e := strconv.Atoi(r[col["token_position"]])
		if e != nil || pos != expected {
			return metadata{}, "", fmt.Errorf("non-contiguous token_position at row %d", expected+2)
		}
		m.Currier = append(m.Currier, normalized(r[col["currier"]]))
		m.Hand = append(m.Hand, normalized(r[col["hand"]]))
		m.Folio = append(m.Folio, normalized(r[col["folio"]]))
		expected++
	}
	if err := s.Err(); err != nil {
		return metadata{}, "", err
	}
	return m, fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func normalized(s string) string {
	if s == "?" || s == "null" {
		return ""
	}
	return s
}

func buildBlocks(m metadata) []block {
	var out []block
	counts := map[string]int{}
	start := -1
	joint := ""
	flush := func(end int) {
		if start < 0 {
			return
		}
		parts := strings.SplitN(joint, "/", 2)
		idx := counts[joint]
		counts[joint]++
		out = append(out, block{ID: joint + "#" + strconv.Itoa(idx), Joint: joint, Currier: parts[0], Hand: parts[1], Index: idx, Start: start, End: end})
		start = -1
	}
	for i := range m.Currier {
		j := ""
		if m.Currier[i] != "" && m.Hand[i] != "" {
			j = m.Currier[i] + "/" + m.Hand[i]
		}
		if start >= 0 && j == joint {
			continue
		}
		flush(i)
		if j != "" {
			start, joint = i, j
		}
	}
	flush(len(m.Currier))
	return out
}

type assignmentKey struct {
	joint             string
	block, start, end int
}

func readAssignments(path string, size, k int) (map[assignmentKey]int, map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64<<10), 8<<20)
	if !s.Scan() {
		return nil, nil, fmt.Errorf("empty assignments: %s", path)
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, v := range h {
		col[v] = i
	}
	for _, n := range []string{"joint_class", "block_index", "abs_start", "abs_end", "window_size", "residual_cluster"} {
		if _, ok := col[n]; !ok {
			return nil, nil, fmt.Errorf("assignments missing %q", n)
		}
	}
	out := map[assignmentKey]int{}
	classes := map[string]bool{}
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		ws, _ := strconv.Atoi(r[col["window_size"]])
		if ws != size {
			continue
		}
		bi, _ := strconv.Atoi(r[col["block_index"]])
		st, _ := strconv.Atoi(r[col["abs_start"]])
		en, _ := strconv.Atoi(r[col["abs_end"]])
		cl, _ := strconv.Atoi(r[col["residual_cluster"]])
		if cl < 0 || cl >= k {
			return nil, nil, fmt.Errorf("assignment cluster %d outside K=%d", cl, k)
		}
		key := assignmentKey{r[col["joint_class"]], bi, st, en}
		if _, dup := out[key]; dup {
			return nil, nil, fmt.Errorf("duplicate assignment for %+v", key)
		}
		out[key] = cl
		classes[key.joint] = true
	}
	if err := s.Err(); err != nil {
		return nil, nil, err
	}
	if len(out) == 0 {
		return nil, nil, fmt.Errorf("no assignments for frozen window_size=%d", size)
	}
	return out, classes, nil
}

func readFrozenOriginalNMI(path string, size, k int) (currier, hand float64, err error) {
	f, e := os.Open(path)
	if e != nil {
		return 0, 0, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return 0, 0, fmt.Errorf("empty association: %s", path)
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, v := range h {
		col[v] = i
	}
	for _, n := range []string{"window_size", "k", "method", "representation", "metadata", "original_nmi"} {
		if _, ok := col[n]; !ok {
			return 0, 0, fmt.Errorf("association missing %q", n)
		}
	}
	found := map[string]bool{}
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		ws, _ := strconv.Atoi(r[col["window_size"]])
		kk, _ := strconv.Atoi(r[col["k"]])
		if ws != size || kk != k || r[col["method"]] != "k_medoids" || r[col["representation"]] != "raw" {
			continue
		}
		v, e := strconv.ParseFloat(r[col["original_nmi"]], 64)
		if e != nil {
			return 0, 0, e
		}
		switch r[col["metadata"]] {
		case "currier":
			currier = v
			found["currier"] = true
		case "hand":
			hand = v
			found["hand"] = true
		}
	}
	if e := s.Err(); e != nil {
		return 0, 0, e
	}
	if !found["currier"] || !found["hand"] {
		return 0, 0, fmt.Errorf("frozen association lacks raw k-medoids window=%d K=%d baseline", size, k)
	}
	return currier, hand, nil
}

type split struct{ train, test []block }

func folds(bs []block) []split {
	if len(bs) >= 2 {
		out := make([]split, len(bs))
		for i := range bs {
			for j, b := range bs {
				if i != j {
					out[i].train = append(out[i].train, b)
				}
			}
			out[i].test = []block{bs[i]}
		}
		return out
	}
	if len(bs) == 1 {
		b := bs[0]
		if b.len() < 6 {
			return nil
		}
		step := b.len() / 3
		out := make([]split, 0, 3)
		for i := 0; i < 3; i++ {
			s, e := b.Start+i*step, b.Start+(i+1)*step
			if i == 2 {
				e = b.End
			}
			tb := b
			tb.Start = s
			tb.End = e
			sp := split{test: []block{tb}}
			if s > b.Start {
				x := b
				x.End = s
				sp.train = append(sp.train, x)
			}
			if e < b.End {
				x := b
				x.Start = e
				sp.train = append(sp.train, x)
			}
			out = append(out, sp)
		}
		return out
	}
	return nil
}

type rawWindow struct {
	profile    sparse
	b          block
	start, end int
}

func windowsFor(tokens []string, bs []block, size int) []rawWindow {
	var out []rawWindow
	for _, b := range bs {
		if b.len() < size {
			continue
		}
		for _, w := range globalregime.BuildWindows(tokens[b.Start:b.End], size, 0) {
			p := sparse{}
			for tok, v := range w.Distribution() {
				p[tok] = v
			}
			out = append(out, rawWindow{p, b, b.Start + w.Start, b.Start + w.End})
		}
	}
	return out
}

func meanProfiles(ws []rawWindow) sparse {
	m := sparse{}
	if len(ws) == 0 {
		return m
	}
	for _, w := range ws {
		for x, v := range w.profile {
			m[x] += v
		}
	}
	for x := range m {
		m[x] /= float64(len(ws))
	}
	return m
}

func subtract(a, b sparse) sparse {
	r := make(sparse, len(a)+len(b))
	for x, v := range a {
		r[x] = v - b[x]
	}
	for x, v := range b {
		if _, ok := a[x]; !ok {
			r[x] = -v
		}
	}
	return r
}

func loadWindows(tokens []string, m metadata, bs []block, assignments map[assignmentKey]int, eligible map[string]bool, size int) ([]window, []foldDiagnostic, error) {
	byClass := map[string][]block{}
	for _, b := range bs {
		if eligible[b.Joint] {
			byClass[b.Joint] = append(byClass[b.Joint], b)
		}
	}
	classes := make([]string, 0, len(byClass))
	for c := range byClass {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	var out []window
	var diagnostics []foldDiagnostic
	for _, c := range classes {
		for fi, sp := range folds(byClass[c]) {
			train, test := windowsFor(tokens, sp.train, size), windowsFor(tokens, sp.test, size)
			if len(train) == 0 || len(test) == 0 {
				continue
			}
			mu := meanProfiles(train)
			whitener := fitWhitening(rawProfiles(train))
			trainResidual := make([]sparse, len(train))
			for i, w := range train {
				trainResidual[i] = subtract(w.profile, mu)
			}
			testResidual := make([]sparse, len(test))
			for i, w := range test {
				testResidual[i] = subtract(w.profile, mu)
			}
			diagnostics = append(diagnostics, foldDiagnostic{size, fi, c, len(train), len(test), normOf(meanSparse(trainResidual)), normOf(meanSparse(testResidual))})
			for i, rw := range test {
				key := assignmentKey{c, rw.b.Index, rw.start, rw.end}
				label, ok := assignments[key]
				if !ok {
					return nil, nil, fmt.Errorf("existing assignments lack reconstructed window %s block %d [%d,%d)", c, rw.b.Index, rw.start, rw.end)
				}
				mid := (rw.start + rw.end - 1) / 2
				out = append(out, window{Currier: rw.b.Currier, Hand: rw.b.Hand, Joint: c, Folio: m.Folio[mid], Block: rw.b.ID, BlockIndex: rw.b.Index, Start: rw.start, End: rw.end, Fold: fi, PhysicalStart: rw.b.Start, PhysicalEnd: rw.b.End, Raw: rw.profile, Residual: testResidual[i], Whitened: whitener.apply(rw.profile), ExistingCluster: label})
			}
		}
	}
	if len(out) != len(assignments) {
		return nil, nil, fmt.Errorf("reconstructed %d windows but frozen assignments contain %d", len(out), len(assignments))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Start == out[j].Start {
			return out[i].Joint < out[j].Joint
		}
		return out[i].Start < out[j].Start
	})
	return out, diagnostics, nil
}

func rawProfiles(ws []rawWindow) []sparse {
	out := make([]sparse, len(ws))
	for i, w := range ws {
		out[i] = w.profile
	}
	return out
}
