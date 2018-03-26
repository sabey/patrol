package main

import (
	"path/filepath"
)

func IsAppKey(key string) bool {
	if len(key) == 0 ||
		len(key) > 32 {
		return false
	}
	for _, r := range key {
		if r < '0' ||
			r > '9' && r < 'A' ||
			r > 'Z' && r < 'a' ||
			r > 'z' {
			return false
		}
	}
	return true
}
func IsPathClean(path string) bool {
	if path == "" ||
		path == "." {
		// no root directories
		return false
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." {
		// no current root directories
		return false
	}
	if cleaned == path {
		// directories match
		return true
	}
	// we don't want to accept // which gets cleaned to / as a clean input
	// this would mean we're comparing // as // which we don't want
	if cleaned != "/" &&
		cleaned+"/" == path {
		// paths are allowed to end with a delimiter
		return true
	}
	// does not match
	return false
}
