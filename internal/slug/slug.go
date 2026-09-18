package slug

import (
	"strings"
	"unicode"
)

// Separator constants.
const (
	SepHyphen     = "-"
	SepUnderscore = "_"
	SepDot        = "."
)

// Options controls slug generation behaviour.
type Options struct {
	Separator string // separator between words (default: -)
	Lowercase bool   // convert to lowercase (default: true)
	MaxLength int    // max slug length, 0 = unlimited
}

// DefaultOptions returns the standard slug options.
func DefaultOptions() Options {
	return Options{
		Separator: SepHyphen,
		Lowercase: true,
		MaxLength: 0,
	}
}

// Generate creates a URL-safe slug from the given text.
func Generate(text string, opts Options) string {
	if opts.Separator == "" {
		opts.Separator = SepHyphen
	}

	// Step 1: Unicode transliteration — convert accented chars to ASCII.
	text = transliterate(text)

	// Step 2: Case conversion.
	if opts.Lowercase {
		text = strings.ToLower(text)
	}

	// Step 3: Replace any non-alphanumeric character with separator.
	var sb strings.Builder
	prevSep := false
	for _, r := range text {
		if isAlnum(r) {
			sb.WriteRune(r)
			prevSep = false
		} else if !prevSep && sb.Len() > 0 {
			sb.WriteString(opts.Separator)
			prevSep = true
		}
	}

	result := sb.String()

	// Trim trailing separator.
	result = strings.TrimRight(result, opts.Separator)

	// Step 4: Truncate to max length (on separator boundary).
	if opts.MaxLength > 0 && len(result) > opts.MaxLength {
		result = result[:opts.MaxLength]
		// Trim back to the last separator boundary to avoid partial words.
		if idx := strings.LastIndex(result, opts.Separator); idx >= 0 {
			result = result[:idx]
		}
		result = strings.TrimRight(result, opts.Separator)
	}

	return result
}

// Validate checks whether a string is a valid slug.
// A valid slug contains only alphanumeric characters and the separator,
// does not start or end with a separator, and has no consecutive separators.
func Validate(slug string, separator string) bool {
	if slug == "" {
		return false
	}
	if separator == "" {
		separator = SepHyphen
	}

	// Must not start or end with separator.
	if strings.HasPrefix(slug, separator) || strings.HasSuffix(slug, separator) {
		return false
	}

	// Must not contain consecutive separators.
	if strings.Contains(slug, separator+separator) {
		return false
	}

	// All characters must be alphanumeric or the separator.
	for _, r := range slug {
		if !isAlnum(r) && !strings.ContainsRune(separator, r) {
			return false
		}
	}

	return true
}

// isAlnum returns true for ASCII alphanumeric characters.
func isAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// transliterate converts common accented/unicode characters to their ASCII equivalents.
// This is a hand-curated mapping covering Latin, Greek, and Cyrillic scripts.
func transliterate(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r < 128 {
			sb.WriteRune(r)
			continue
		}
		if rep, ok := transliterationTable[r]; ok {
			sb.WriteString(rep)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Keep other letters/digits as-is (they'll be handled by isAlnum check).
			sb.WriteRune(r)
		}
		// Non-letter, non-digit unicode chars are dropped (become separators).
	}
	return sb.String()
}
