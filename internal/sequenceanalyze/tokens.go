package sequenceanalyze

import (
	"bufio"
	"os"
	"strings"
)

// ReadTokens is the shared canonical corpus loader used by corpus-level
// analyses. It deliberately uses the same Go strings.Fields contract as
// sequence-analyze and preserves token order while ignoring physical lines.
func ReadTokens(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var tokens []string
	for s.Scan() {
		tokens = append(tokens, strings.Fields(s.Text())...)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}
