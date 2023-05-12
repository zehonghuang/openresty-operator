package utils

import (
	"regexp"
	"strings"
)

func ValidateLocationPath(path string) (bool, string) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false, "path cannot be empty"
	}

	// Regex match: ~ or ~*
	if strings.HasPrefix(trimmed, "~") {
		if !regexp.MustCompile(`^~\*?\s+.+`).MatchString(trimmed) {
			return false, "invalid regex path format"
		}
		re := strings.TrimSpace(strings.TrimPrefix(trimmed, "~"))
		re = strings.TrimSpace(strings.TrimPrefix(re, "*"))
		if _, err := regexp.Compile(re); err != nil {
			return false, "invalid regular expression"
		}
		return true, ""
	}

	// Exact match: =
	if strings.HasPrefix(trimmed, "=") {
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "="))
		if !strings.HasPrefix(rest, "/") {
			return false, "exact path must start with '/'"
		}
		return true, ""
	}

	// Strong prefix match: ^~
	if strings.HasPrefix(trimmed, "^~") {
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "^~"))
		if !strings.HasPrefix(rest, "/") {
			return false, "prefix path must start with '/'"
		}
		return true, ""
	}

	// Normal prefix match
	if !strings.HasPrefix(trimmed, "/") {
		return false, "path must start with '/'"
	}

	if strings.Contains(trimmed, " ") {
		return false, "path should not contain spaces"
	}

	if len(trimmed) > 256 {
		return false, "path too long"
	}

	return true, ""
}
