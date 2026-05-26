package util

import (
	"fmt"
	"strings"
)

// FormatSize converts bytes to a human-readable string (B, KB, MB, GB).
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatNumber formats an int64 with comma separators (e.g. 1,234,567).
func FormatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		// handle negative: remove leading minus, format, re-add
		s = fmt.Sprintf("%d", -n)
		return "-" + insertCommas(s)
	}
	return insertCommas(s)
}

// insertCommas inserts commas every three digits from the right.
func insertCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	rem := n % 3
	if rem > 0 {
		b.WriteString(s[:rem])
		if n > rem {
			b.WriteByte(',')
		}
	}
	for i := rem; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteByte(',')
		}
	}
	return b.String()
}

// EscapeCSV wraps s in double-quotes if it contains a comma, newline, or
// double-quote. Any double-quotes within the string are doubled ("").
func EscapeCSV(s string) string {
	if strings.ContainsAny(s, ",\n\r\"") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
