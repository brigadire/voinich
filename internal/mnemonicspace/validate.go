package mnemonicspace

import "fmt"

func ValidateRegistry(authority Authority, specs []MechanismSpec) error {
	seen := map[string]bool{}
	for _, spec := range specs {
		if seen[spec.ID] {
			return fmt.Errorf("duplicate mechanism %q", spec.ID)
		}
		seen[spec.ID] = true
		if err := ValidateMechanism(authority, spec); err != nil {
			return fmt.Errorf("%s: %w", spec.ID, err)
		}
	}
	return nil
}

func ValidateMechanism(authority Authority, spec MechanismSpec) error {
	if spec.ID == "" || spec.Version == "" || spec.Family == "" {
		return fmt.Errorf("missing identifier fields")
	}
	if spec.runtime == kindUnknown {
		return fmt.Errorf("unknown runtime kind")
	}
	if !authority.FutureStatus[spec.Status] {
		return fmt.Errorf("status %q is not permitted by task80", spec.Status)
	}
	if spec.Status == StatusForbidden || spec.Status == StatusReferenceOnly {
		return fmt.Errorf("non-runnable status %q", spec.Status)
	}
	if len(spec.Parameters) == 0 {
		return fmt.Errorf("missing parameter sets")
	}
	if !spec.Serialization.ExcludesHiddenState {
		return fmt.Errorf("observable serialization must exclude hidden state")
	}
	for _, carrier := range spec.Carriers.Retrieve {
		if carrier == CarrierMessage {
			return fmt.Errorf("retrieval must not depend on message carrier")
		}
	}
	for _, model := range spec.SourceModels {
		if authority.ReferenceOnly[model] {
			return fmt.Errorf("reference-only model %s cannot be runnable", model)
		}
		if authority.Excluded[model] {
			return fmt.Errorf("excluded model %s cannot be runnable", model)
		}
		if model != "" {
			if _, ok := authority.Models[model]; !ok {
				return fmt.Errorf("unknown source model %s", model)
			}
		}
	}
	for _, step := range append(append([]OperationInvocation{}, spec.Encoding...), spec.Retrieval...) {
		op, ok := authority.Operations[step.Operation]
		if !ok {
			return fmt.Errorf("unknown task80 operation %s", step.Operation)
		}
		if len(step.Inputs) != len(op.InputTypes) {
			return fmt.Errorf("%s input arity mismatch", step.Operation)
		}
		for i, input := range step.Inputs {
			if string(input.Type) != op.InputTypes[i] {
				return fmt.Errorf("%s input %d has type %s, want %s", step.Operation, i, input.Type, op.InputTypes[i])
			}
		}
		if len(step.Outputs) != len(op.OutputTypes) {
			return fmt.Errorf("%s output arity mismatch", step.Operation)
		}
		for i, output := range step.Outputs {
			if string(output.Type) != op.OutputTypes[i] {
				return fmt.Errorf("%s output %d has type %s, want %s", step.Operation, i, output.Type, op.OutputTypes[i])
			}
		}
	}
	if err := validateCueConversions(spec); err != nil {
		return err
	}
	if err := validateFamilyShape(spec); err != nil {
		return err
	}
	for _, param := range spec.Parameters {
		if param.ID == "" {
			return fmt.Errorf("parameter set without id")
		}
		if !param.Frozen {
			return fmt.Errorf("parameter set %s must be frozen", param.ID)
		}
	}
	return nil
}

func validateCueConversions(spec MechanismSpec) error {
	outputs := map[string]DomainType{}
	associateCueInputs := map[string]bool{}
	for _, step := range append(append([]OperationInvocation{}, spec.Encoding...), spec.Retrieval...) {
		for _, output := range step.Outputs {
			outputs[output.Name] = output.Type
		}
		if step.Operation == "associate" {
			for _, input := range step.Inputs {
				if input.Type == TypeCue {
					associateCueInputs[input.Name] = true
				}
			}
		}
	}
	conversions := map[string]CueConversion{}
	for _, conversion := range spec.CueConversions {
		if conversion.InputName == "" || conversion.OutputName == "" || conversion.Rule == "" {
			return fmt.Errorf("incomplete cue conversion")
		}
		if conversion.From != TypeSymbol || conversion.To != TypeCue {
			return fmt.Errorf("cue conversion %s must be Symbol -> Cue", conversion.OutputName)
		}
		if outputs[conversion.InputName] != TypeSymbol {
			return fmt.Errorf("cue conversion %s has no Symbol producer %s", conversion.OutputName, conversion.InputName)
		}
		if _, exists := conversions[conversion.OutputName]; exists {
			return fmt.Errorf("duplicate cue conversion output %s", conversion.OutputName)
		}
		conversions[conversion.OutputName] = conversion
	}
	for cueName := range associateCueInputs {
		if outputs[cueName] == TypeCue {
			continue
		}
		if _, ok := conversions[cueName]; !ok {
			return fmt.Errorf("associate cue input %s has no Cue producer or declared Symbol -> Cue conversion", cueName)
		}
	}
	return nil
}

func validateFamilyShape(spec MechanismSpec) error {
	requires := map[Family][]OperationID{
		FamilyF01:                    {"rotate", "align", "traverse", "compose"},
		FamilyF08:                    {"place", "traverse", "compose"},
		FamilyF11:                    {"place", "index", "select"},
		FamilyF12:                    {"repeat", "signal", "associate"},
		FamilyRestrictedRotation:     {"rotate", "index", "select"},
		FamilyRestrictedAssociation:  {"place", "select", "associate"},
		FamilyGenericLiteral:         {"place", "traverse", "compose"},
		FamilyGenericCyclic:          {"repeat", "signal"},
		FamilyGenericIndexed:         {"place", "index", "select"},
		FamilyGenericCue:             {"repeat", "signal", "associate"},
		FamilyGenericAmbiguous:       {"place", "select", "associate"},
		FamilyNegativeConvention:     {"rotate", "align", "traverse", "compose"},
		FamilyNegativePath:           {"place", "traverse", "compose"},
		FamilyNegativeCueAssociation: {"repeat", "signal", "associate"},
		FamilyNegativeIndexMapping:   {"place", "index", "select"},
	}
	needed := requires[spec.Family]
	if len(needed) == 0 {
		return fmt.Errorf("unknown family %s", spec.Family)
	}
	seen := map[OperationID]bool{}
	for _, step := range append(spec.Encoding, spec.Retrieval...) {
		seen[step.Operation] = true
	}
	for _, op := range needed {
		if !seen[op] {
			return fmt.Errorf("family %s is missing %s", spec.Family, op)
		}
	}
	return nil
}
