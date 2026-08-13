package clustermetadataglobal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"zcore.dev/voinich/internal/metadatavalidation"
)

// readDiscoveryTokenCount reads the frozen corpus token count that discovery
// itself recorded. Frozen windows may be a sliding, overlapping sweep whose
// last window does not reach the final token, so the search space's own
// window bounds cannot be used to infer the corpus length; the discovery
// metadata is the single source of truth for it.
func readDiscoveryTokenCount(discoveryDir string) (int, error) {
	path := filepath.Join(discoveryDir, "global_distributional_regimes.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read frozen discovery metadata: %w", err)
	}
	var x struct {
		Parameters struct {
			TokenCount int `yaml:"token_count"`
		} `yaml:"parameters"`
	}
	if err := yaml.Unmarshal(b, &x); err != nil {
		return 0, err
	}
	if x.Parameters.TokenCount <= 0 {
		return 0, fmt.Errorf("frozen discovery metadata %s has no token_count", path)
	}
	return x.Parameters.TokenCount, nil
}

// loadFrozenSpace loads the frozen cluster assignments and validates that
// they cover exactly the prespecified window_size x method x K search space,
// with a shared, consistent window layout per window size. It never
// recomputes windows, clustering or cluster assignments.
func loadFrozenSpace(discoveryDir string) (*frozenSpace, error) {
	path := filepath.Join(discoveryDir, "global_distributional_cluster_assignments.tsv")
	rows, err := metadatavalidation.LoadAssignments(path)
	if err != nil {
		return nil, fmt.Errorf("load frozen cluster assignments: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no frozen cluster assignments in %s", path)
	}
	windows := map[int]map[int]WindowRange{}
	combos := map[comboKey]map[int]int{}
	maxCluster := map[comboKey]int{}
	maxEnd := 0
	for _, a := range rows {
		byIdx := windows[a.WindowSize]
		if byIdx == nil {
			byIdx = map[int]WindowRange{}
			windows[a.WindowSize] = byIdx
		}
		if existing, ok := byIdx[a.Index]; ok {
			if existing.Start != a.Start || existing.End != a.End {
				return nil, fmt.Errorf("inconsistent frozen window bounds for window_size=%d index=%d: (%d,%d) vs (%d,%d)",
					a.WindowSize, a.Index, existing.Start, existing.End, a.Start, a.End)
			}
		} else {
			byIdx[a.Index] = WindowRange{a.Index, a.Start, a.End}
		}
		if a.End > maxEnd {
			maxEnd = a.End
		}
		key := comboKey{a.WindowSize, a.Method, a.K}
		if combos[key] == nil {
			combos[key] = map[int]int{}
		}
		combos[key][a.Index] = a.Cluster
		if a.Cluster+1 > maxCluster[key] {
			maxCluster[key] = a.Cluster + 1
		}
	}
	fs := &frozenSpace{Windows: map[int][]WindowRange{}, Combos: map[comboKey]comboData{}, N: maxEnd}
	for ws, byIdx := range windows {
		idxs := make([]int, 0, len(byIdx))
		for i := range byIdx {
			idxs = append(idxs, i)
		}
		sort.Ints(idxs)
		for i, idx := range idxs {
			if idx != i {
				return nil, fmt.Errorf("frozen windows for window_size=%d are not contiguous from 0 (got index %d at position %d)", ws, idx, i)
			}
		}
		ranges := make([]WindowRange, len(idxs))
		for i, idx := range idxs {
			ranges[i] = byIdx[idx]
		}
		fs.Windows[ws] = ranges
	}
	for _, ws := range WindowSizes {
		ranges, ok := fs.Windows[ws]
		if !ok {
			return nil, fmt.Errorf("frozen search space is missing window_size=%d entirely", ws)
		}
		for _, m := range Methods {
			for _, k := range ksRange() {
				key := comboKey{ws, m, k}
				byIdx, ok := combos[key]
				if !ok {
					return nil, fmt.Errorf("frozen search space is missing combo window_size=%d method=%s k=%d", ws, m, k)
				}
				cluster := make([]int, len(ranges))
				for i, r := range ranges {
					c, ok := byIdx[r.Index]
					if !ok {
						return nil, fmt.Errorf("frozen search space is missing window_index=%d for combo window_size=%d method=%s k=%d", r.Index, ws, m, k)
					}
					cluster[i] = c
				}
				fs.Combos[key] = comboData{Cluster: cluster, NumClusters: maxCluster[key]}
			}
		}
	}
	return fs, nil
}

// loadTokenLabels reads only the columns needed from the already-produced,
// frozen-aligned token_metadata_map.tsv. It never re-runs IVTFF parsing or
// alignment; those remain the responsibility of metadata-validate.
func loadTokenLabels(path string) (currier, hand []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	if !s.Scan() {
		if e := s.Err(); e != nil {
			return nil, nil, e
		}
		return nil, nil, fmt.Errorf("empty token metadata map: %s", path)
	}
	header := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	for _, need := range []string{"token_position", "currier", "hand"} {
		if _, ok := col[need]; !ok {
			return nil, nil, fmt.Errorf("token metadata map %s is missing required column %q", path, need)
		}
	}
	expect := 0
	for s.Scan() {
		row := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
		posRaw := row[col["token_position"]]
		pos, e := strconv.Atoi(posRaw)
		if e != nil {
			return nil, nil, fmt.Errorf("token metadata map %s: non-integer token_position %q", path, posRaw)
		}
		if pos != expect {
			return nil, nil, fmt.Errorf("token metadata map %s: non-contiguous token_position %d, expected %d", path, pos, expect)
		}
		expect++
		currier = append(currier, normalizeLabel(row[col["currier"]]))
		hand = append(hand, normalizeLabel(row[col["hand"]]))
	}
	if e := s.Err(); e != nil {
		return nil, nil, e
	}
	return currier, hand, nil
}

func normalizeLabel(v string) string {
	if v == "" || v == "?" || v == "null" {
		return ""
	}
	return v
}
