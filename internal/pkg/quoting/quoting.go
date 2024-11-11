// Package quoting provides helpers to determine if a string needs to be quoted.
package quoting

import "unicode/utf8"

// IsNeeded returns true if the string needs to be quoted.
func IsNeeded(s string) bool {
	looksLikeNumber := true
	hasDigits := false
	nDots := 0

	for _, char := range s {
		switch char {
		case '.':
			nDots++
		case '=', '"', ' ', utf8.RuneError:
			return true
		default:
			if char < ' ' {
				return true
			}

			if isDigit(char) {
				hasDigits = true
			} else {
				looksLikeNumber = false
			}
		}
	}

	return looksLikeNumber && hasDigits && nDots <= 1
}

// ---

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
