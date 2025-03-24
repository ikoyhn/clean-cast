package common

import (
	"strings"
	"unicode"
)

func IsValidFilename(filename string) bool {
	for _, c := range filename {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '.' && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func IsValidParam(param string) bool {
	if strings.Contains(param, "/") || strings.Contains(param, "\\") || strings.Contains(param, "..") {
		return false
	}
	return true
}

func IsValidID(id string) bool {
	for _, c := range id {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
