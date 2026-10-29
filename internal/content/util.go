package content

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func capitalize(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[n:]
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToLower(r)) + s[n:]
}

// pick chooses an element deterministically from a pool.
func pick(pool []string, seed int64) string {
	if len(pool) == 0 {
		return ""
	}
	i := seed % int64(len(pool))
	if i < 0 {
		i += int64(len(pool))
	}
	return pool[i]
}

// hash is a small stable string hash (FNV-1a) used to vary wording.
func hash(parts ...string) int64 {
	var h uint64 = 14695981039346656037
	for _, p := range parts {
		for i := 0; i < len(p); i++ {
			h ^= uint64(p[i])
			h *= 1099511628211
		}
		h ^= '|'
		h *= 1099511628211
	}
	return int64(h >> 1)
}

func join(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}
