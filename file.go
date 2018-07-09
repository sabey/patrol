package patrol

import (
	"fmt"
	"os"
	p "path"
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
	path = p.Clean(path)
	if path == "." ||
		path == ".." {
		// invalid
		return nil, ERR_PATH_INVALID
	}
	// path is usable
	// we have to open our file for writing, we don't care about reading
	// we're going to truncate our file if it exists
	// if our file does not exist we're going to delete it
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
