package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParsePrice extracts numeric price from Indonesian formatted string
// Examples: "Rp 1.000.000.000", "Rp 500 Juta"
func ParsePrice(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty price string")
	}

	// Remove "Rp", spaces, and dots
	s = strings.ReplaceAll(s, "Rp", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.TrimSpace(s)

	// Handle "Juta" (million) and "Miliar" (billion)
	multiplier := 1.0
	if strings.Contains(strings.ToLower(s), "miliar") || strings.Contains(strings.ToLower(s), "m") {
		multiplier = 1000000000.0
		s = regexp.MustCompile(`(?i)(miliar|m)`).ReplaceAllString(s, "")
	} else if strings.Contains(strings.ToLower(s), "juta") {
		multiplier = 1000000.0
		s = regexp.MustCompile(`(?i)juta`).ReplaceAllString(s, "")
	}

	s = strings.TrimSpace(s)

	// Extract first number
	re := regexp.MustCompile(`[\d,]+`)
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no numeric value found in: %s", s)
	}

	// Remove commas
	match = strings.ReplaceAll(match, ",", "")

	// Parse number
	num, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing number: %w", err)
	}

	return num * multiplier, nil
}

// ParseInt extracts integer from string
func ParseInt(s string) int {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}

	num, err := strconv.Atoi(match)
	if err != nil {
		return 0
	}

	return num
}

// ParseFloat extracts float from string
func ParseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")

	re := regexp.MustCompile(`[\d.]+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}

	num, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}

	return num
}

// CleanText removes extra whitespace and trims
func CleanText(s string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// MakeAbsoluteURL converts relative URL to absolute using base URL
func MakeAbsoluteURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}

	if strings.HasPrefix(relativeURL, "/") {
		// Extract domain from base URL
		re := regexp.MustCompile(`^(https?://[^/]+)`)
		match := re.FindString(baseURL)
		if match != "" {
			return match + relativeURL
		}
	}

	return relativeURL
}
