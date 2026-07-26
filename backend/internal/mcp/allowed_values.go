package mcp

import "fmt"

// validateAllowedValues checks raw against valid, de-duplicating as it goes.
// An empty raw means "all of valid." fieldName only names the field in the
// error message (e.g. "record_type", "status").
func validateAllowedValues(raw []string, valid []string, fieldName string) ([]string, error) {
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
			return nil, fmt.Errorf("invalid %s: %q", fieldName, s)
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out, nil
}
