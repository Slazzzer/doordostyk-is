package handler

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	phoneRE = regexp.MustCompile(`^\+7\d{10}$`)
)

func isStrongPassword(v string) bool {
	if len(v) < 8 || len(v) > 72 {
		return false
	}
	const specials = `!@#$%^&*()_+-=[]{};':"\|,.<>/?` + "`" + `~`
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	for _, r := range v {
		if r > unicode.MaxASCII {
			return false
		}
		if unicode.IsUpper(r) {
			hasUpper = true
			continue
		}
		if unicode.IsLower(r) {
			hasLower = true
			continue
		}
		if unicode.IsDigit(r) {
			hasDigit = true
			continue
		}
		if strings.ContainsRune(specials, r) {
			hasSpecial = true
			continue
		}
		return false
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}
