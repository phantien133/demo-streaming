package stringsutil

import "strings"

// FirstNonEmpty returns the first non-empty string after TrimSpace.
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

