// Package validation holds shared request-validation helpers used by internal/service.
package validation

// SizeTier is a named text-length limit shared across all request validation,
// so callers pick a tier (e.g. Small) instead of a hand-picked character count.
type SizeTier int

const (
	SuperSmall SizeTier = iota
	Small
	Medium
	Large
	ExtraLarge
)

// maxLengths maps each SizeTier to its character limit.
var maxLengths = map[SizeTier]int{
	SuperSmall: 20,
	Small:      50,
	Medium:     150,
	Large:      500,
	ExtraLarge: 5000,
}

// MaxLength returns the maximum character count allowed for the tier.
func (t SizeTier) MaxLength() int {
	return maxLengths[t]
}

// TooLong reports whether value exceeds the tier's character limit (rune-counted,
// so multi-byte characters count as one each rather than by their UTF-8 byte length).
func TooLong(value string, tier SizeTier) bool {
	return len([]rune(value)) > tier.MaxLength()
}
