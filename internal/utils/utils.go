package utils

import (
	"os"
	"strings"
)

// CanonicalSQLKey Line to Hump
// hello_world => HelloWorld
func CanonicalSQLKey(s string) string {
	if s == "" {
		return s
	}

	s = strings.Trim(s, "_")
	ss := strings.Split(s, "_")

	result := ""

	for _, v := range ss {
		result += strings.ToUpper(string(v[0])) + strings.ToLower(v[1:])
	}
	return result
}

// Exists checks if a file or directory exists.
func Exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if !os.IsNotExist(err) {
		return false
	}
	return false
}
