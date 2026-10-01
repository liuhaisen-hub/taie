package utils

func TruncateWithEllipsis(s string, maxRune int) string {
	runes := []rune(s)
	if len(runes) > maxRune {
		return string(runes[:maxRune-1]) + "…"
	}
	return s
}
