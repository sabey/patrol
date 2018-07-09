package patrol

import (
	"path"
	"testing"
)

func TestFile(t *testing.T) {
	// open file
	f, err := OpenFile("unittest/test.log")
	if err != nil {
		t.Fatalf("failed to open test.log: \"%s\"\n", err)
		return
	}
	// close file
	f.Close()
	// open directory
	f, err = OpenFile("unittest")
	if err == nil {
		t.Fatal("opened unittest `directory` as a file!")
		return
	}
	if f != nil {
		f.Close()
		t.Fatal("unittest `file` was non nil!?")
		return
	}
}
func TestPath(t *testing.T) {
	// current
	for _, s := range []string{
		".",
		"./",
		"./.",
		"a/..",
	} {
		if r := path.Clean(s); r != "." {
			t.Fatalf("path: \"%s\" did not match current directory, was: \"%s\"\n", s, r)
			return
		}
		if f, err := OpenFile(s); err != ERR_PATH_INVALID {
			f.Close()
			t.Fatalf("current path: \"%s\" was valid?! err: \"%s\"\n", s, err)
			return
		}
	}
	// parent
	for _, s := range []string{
		"..",
		"../",
		"./..",
	} {
		if r := path.Clean(s); r != ".." {
			t.Fatalf("path: \"%s\" did not match parent directory, was: \"%s\"\n", s, r)
			return
		}
		if f, err := OpenFile(s); err != ERR_PATH_INVALID {
			f.Close()
			t.Fatalf("parent path: \"%s\" was valid?! err: \"%s\"\n", s, err)
			return
		}
	}
	// neither
	for _, s := range []string{
		"/.",
		"/..",
		"a/.",
		"/a/.",
		"/a/..",
	} {
		if r := path.Clean(s); r == "." || r == ".." {
			t.Fatalf("path: \"%s\" did not match either `.` or `..`, was: \"%s\"\n", s, r)
			return
		}
	}
}
