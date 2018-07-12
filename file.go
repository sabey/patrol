package patrol

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	ERR_PATH_INVALID = fmt.Errorf("Path can NOT be current or parent Directory!")
	ERR_DIRECTORY    = fmt.Errorf("Path was a Directory")
)

func OpenFile(
	path string,
) (
	*os.File,
	error,
) {
	// validate csv path
	// path.Clean will return an absolute path with one exception
	// it can return "." and ".."
	// if either of these values are returned they are useless to us, we can't write to the current or parent directories
	path = filepath.Clean(path)
	if path == "." ||
		path == ".." {
		// invalid
		return nil, ERR_PATH_INVALID
	}
	// path is usable
	// we need to open our file for writing, we don't care about writing
	// we don't care if our file already exists, append if it does!
	// we also need to create our file if it does not exist
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// failed to open
		return nil, err
	}
	// returned file handler must be closed
	// if an error occurs we must close this handler
	// check if our "file" is a directory
	i, err := f.Stat()
	if err != nil {
		// failed to stat
		f.Close()
		return nil, err
	}
	if i.IsDir() {
		// this is a directory, we can't use this!
		f.Close()
		return nil, ERR_DIRECTORY
	}
	// file is safe to use
	// returned file handler MUST be closed
	return f, nil
}
