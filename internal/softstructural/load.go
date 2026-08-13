package softstructural

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/structuralreliability"
)

type dictionaryEntry struct {
	Token     string `yaml:"token"`
	Count     int    `yaml:"count"`
	Positions []struct {
		Position int `yaml:"position"`
		Count    int `yaml:"count"`
	} `yaml:"position_in_string"`
	Before []struct {
		Token string `yaml:"token"`
		Count int    `yaml:"count"`
	} `yaml:"word_before"`
	After []struct {
		Token string `yaml:"token"`
		Count int    `yaml:"count"`
	} `yaml:"word_after"`
}
type analysisEntry struct {
	Token string `yaml:"token"`
	Count int    `yaml:"count"`
}
type reliabilityInput struct {
	Curves         structuralreliability.ReliabilityCurves `yaml:"reliability_curves"`
	ReferencePairs []struct {
		TokenA    string   `yaml:"token_a"`
		TokenB    string   `yaml:"token_b"`
		Bootstrap *float64 `yaml:"bootstrap_probability_above_070"`
	} `yaml:"reference_pairs"`
}

func readYAML(path string, destination any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, destination)
}

func loadInputs(config Config) (dataset, reliabilityInput, error) {
	var dictionary []dictionaryEntry
	if err := readYAML(config.DictionaryPath, &dictionary); err != nil {
		return dataset{}, reliabilityInput{}, fmt.Errorf("read dictionary: %w", err)
	}
	var analysis []analysisEntry
	if err := readYAML(config.AnalysisPath, &analysis); err != nil {
		return dataset{}, reliabilityInput{}, fmt.Errorf("read analysis: %w", err)
	}
	analysisCounts := make(map[string]int, len(analysis))
	for _, item := range analysis {
		analysisCounts[item.Token] = item.Count
	}
	d := dataset{profiles: make(map[string]profilestability.Profile, len(dictionary)), counts: make(map[string]int, len(dictionary))}
	for _, item := range dictionary {
		if item.Token == "" || item.Count < 0 {
			return dataset{}, reliabilityInput{}, fmt.Errorf("invalid dictionary token %q", item.Token)
		}
		if got, ok := analysisCounts[item.Token]; !ok || got != item.Count {
			return dataset{}, reliabilityInput{}, fmt.Errorf("analysis count mismatch for %q", item.Token)
		}
		if _, exists := d.profiles[item.Token]; exists {
			return dataset{}, reliabilityInput{}, fmt.Errorf("duplicate dictionary token %q", item.Token)
		}
		profile := profilestability.Profile{Count: item.Count, Positions: map[int]int{}, Left: map[string]int{}, Right: map[string]int{}}
		for _, value := range item.Positions {
			profile.Positions[value.Position] += value.Count
		}
		for _, value := range item.Before {
			profile.Left[value.Token] += value.Count
		}
		for _, value := range item.After {
			profile.Right[value.Token] += value.Count
		}
		d.profiles[item.Token], d.counts[item.Token] = profile, item.Count
	}
	if len(analysisCounts) != len(d.profiles) {
		return dataset{}, reliabilityInput{}, fmt.Errorf("dictionary/analysis token count mismatch")
	}
	var reliability reliabilityInput
	if err := readYAML(config.ReliabilityPath, &reliability); err != nil {
		return dataset{}, reliabilityInput{}, fmt.Errorf("read reliability: %w", err)
	}
	if len(reliability.Curves.Position) == 0 || len(reliability.Curves.LeftContext) == 0 || len(reliability.Curves.RightContext) == 0 {
		return dataset{}, reliabilityInput{}, fmt.Errorf("reliability curves must not be empty")
	}
	return d, reliability, nil
}
