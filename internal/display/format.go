package display

import "strings"

// FormatStatus formats risk status for display
func FormatStatus(status string) string {
	switch status {
	case "detected":
		return "[DETECTED]"
	case "accepted":
		return "[ACCEPTED]"
	case "resolved":
		return "[RESOLVED]"
	case "archived":
		return "[ARCHIVED]"
	default:
		return "[" + strings.ToUpper(status) + "]"
	}
}

// FormatPriority formats risk score as priority
func FormatPriority(score int) string {
	if score >= 20 {
		return "CRITICAL"
	} else if score >= 15 {
		return "HIGH"
	} else if score >= 10 {
		return "MEDIUM"
	}
	return "LOW"
}

// FormatControlType formats control type for display
func FormatControlType(controlType string) string {
	switch controlType {
	case "preventive":
		return "[PREVENTIVE]"
	case "detective":
		return "[DETECTIVE]"
	case "corrective":
		return "[CORRECTIVE]"
	default:
		return "[" + strings.ToUpper(controlType) + "]"
	}
}

// FormatWeightTier returns a human-readable tier label for a control weight (1-10)
func FormatWeightTier(weight int) string {
	if weight >= 9 {
		return "Critical"
	} else if weight >= 7 {
		return "Required"
	} else if weight >= 5 {
		return "Important"
	} else if weight >= 3 {
		return "Recommended"
	}
	return "Advisory"
}

// FormatCategory formats a snake_case category into Title Case
func FormatCategory(category string) string {
	words := strings.Split(category, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// FormatValidationStatus formats validation status for display
func FormatValidationStatus(status string) string {
	switch status {
	case "analyst_validated":
		return "[VALIDATED]"
	case "auto_extracted":
		return "[AUTO]"
	case "contradicted":
		return "[CONTRADICTED]"
	default:
		return "[" + strings.ToUpper(status) + "]"
	}
}

// FormatEvidenceStatus formats evidence status for display
func FormatEvidenceStatus(status string) string {
	switch status {
	case "not_configured":
		return "[NOT CONFIGURED]"
	case "configured":
		return "[CONFIGURED]"
	case "sample":
		return "[SAMPLE]"
	case "verified":
		return "[VERIFIED]"
	default:
		return "[" + strings.ToUpper(status) + "]"
	}
}

// TruncateText truncates text to maxLen with ellipsis
func TruncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// WrapText wraps text to a specified width with optional indent
func WrapText(text string, width int, indent string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width-len(indent) {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n"+indent)
}
