package patrol

import (
	"path/filepath"
)

func IsAppServiceID(
	key string,
) bool {
	if len(key) == 0 ||
		len(key) > 63 {
		return false
	}
	// ( 0-9 A-Z a-z - )
	// if these accepted characters are to change, we are unwilling to ever support the character '/' and solo '.' or '..' sequences
	// the key is intended to only be used in a URI path, there are no plans to use this as a label in a hostname
	// if we change this, characters that are not a valid URI path delimiter must be encoded for use in our URL
	if key[0] == '-' ||
		key[len(key)-1] == '-' {
		// we can't accept current directory or parent directory keywords
		return false
	}
	for _, r := range key {
		if r != '-' {
			if r < '0' ||
				r > '9' && r < 'A' ||
				r > 'Z' && r < 'a' ||
				r > 'z' {
				return false
			}
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
func dereference(
	data []byte,
) []byte {
	safe := make([]byte, len(data))
	copy(safe, data)
	return safe
}
