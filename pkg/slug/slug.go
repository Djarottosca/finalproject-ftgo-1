// Package slug generates URL-safe slugs from human-readable names.
package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var nonAlnumRegexp = regexp.MustCompile(`[^a-z0-9]+`)

// Generate builds a URL-safe, unique-enough slug from name.
// ponytail: no DB collision check beyond the unique constraint + random
// suffix. Good enough at this scale; add a retry-on-conflict loop if
// duplicate names become common.
func Generate(name string) string {
	base := nonAlnumRegexp.ReplaceAllString(strings.ToLower(name), "-")
	base = strings.Trim(base, "-")

	suffix := make([]byte, 3)
	_, _ = rand.Read(suffix)

	return base + "-" + hex.EncodeToString(suffix)
}
