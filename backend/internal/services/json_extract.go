package services

import "strings"

// extractJSONFromLLM extracts a JSON object from a string that may contain
// markdown code blocks.
func extractJSONFromLLM(s string) string {
	if idx := strings.Index(s, "```json"); idx >= 0 {
		start := idx + 7
		end := strings.Index(s[start:], "```")
		if end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	if idx := strings.Index(s, "```"); idx >= 0 {
		start := idx + 3
		if nlIdx := strings.Index(s[start:], "\n"); nlIdx >= 0 {
			start += nlIdx + 1
		}
		end := strings.Index(s[start:], "```")
		if end >= 0 {
			candidate := strings.TrimSpace(s[start : start+end])
			if strings.HasPrefix(candidate, "{") {
				return candidate
			}
		}
	}

	if idx := strings.Index(s, "{"); idx >= 0 {
		depth := 0
		for i := idx; i < len(s); i++ {
			switch s[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return s[idx : i+1]
				}
			}
		}
	}

	return ""
}
