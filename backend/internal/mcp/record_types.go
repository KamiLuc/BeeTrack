package mcp

import "fmt"

// validateRecordTypes checks raw against valid, de-duplicating as it goes.
// An empty raw means "all of valid."
func validateRecordTypes(raw []string, valid []string) ([]string, error) {
	if len(raw) == 0 {
		return valid, nil
	}
	validSet := make(map[string]bool, len(valid))
	for _, v := range valid {
		validSet[v] = true
	}
	seen := make(map[string]bool, len(raw))
	var out []string
	for _, s := range raw {
		if !validSet[s] {
			return nil, fmt.Errorf("invalid record_type: %q", s)
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out, nil
}
