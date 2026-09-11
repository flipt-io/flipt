package names

import (
	"strings"
	"unicode"
)

// RefSlug returns name in a form that Git accepts as one path component of a
// reference name (see git-check-ref-format). Each character or sequence that
// Git does not permit is replaced with a hyphen. A name that is already valid
// is returned unchanged, so existing branch names keep their exact form.
//
// Examples:
//
//	"Dev Playground" -> "Dev-Playground"
//	"my/feature"     -> "my-feature"
//	"release..1"     -> "release.-1"
//	"production"     -> "production"
func RefSlug(name string) string {
	var (
		b    strings.Builder
		prev rune
	)

	b.Grow(len(name))

	for _, r := range name {
		switch {
		case r < 0x20 || r == 0x7f || unicode.IsSpace(r):
			// control characters and whitespace
			r = '-'
		case strings.ContainsRune("~^:?*[\\/", r):
			// characters git rejects anywhere in a reference name
			r = '-'
		case r == '.' && (prev == '.' || prev == 0):
			// no ".." and no leading "."
			r = '-'
		case r == '{' && prev == '@':
			// no "@{"
			r = '-'
		}

		b.WriteRune(r)
		prev = r
	}

	s := b.String()

	if s == "@" {
		return "-"
	}

	if base, ok := strings.CutSuffix(s, ".lock"); ok {
		s = base + "-lock"
	}

	if base, ok := strings.CutSuffix(s, "."); ok {
		s = base + "-"
	}

	return s
}
