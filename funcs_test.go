package main

import (
	"fmt"
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestFuncs(t *testing.T) {
	log.Println("TestFuncs")

	fmt.Println("IsAppKey")
	unittest.Equals(t, IsAppKey(""), false)
	unittest.Equals(t, IsAppKey("12345678901234567901234567891234"), true)
	unittest.Equals(t, IsAppKey("123456789012345679012345678912345"), false)

	for i := 0; i <= 0x7F; i++ {
		if i >= '0' && i <= '9' ||
			i >= 'A' && i <= 'Z' ||
			i >= 'a' && i <= 'z' {
			unittest.Equals(t, IsAppKey(string([]byte{byte(i)})), true)
		} else {
			unittest.Equals(t, IsAppKey(string([]byte{byte(i)})), false)
		}
	}

	fmt.Println("IsPathClean")
	// invalid
	unittest.Equals(t, IsPathClean(""), false)
	unittest.Equals(t, IsPathClean("."), false)
	unittest.Equals(t, IsPathClean("./"), false)
	unittest.Equals(t, IsPathClean("a/."), false)
	unittest.Equals(t, IsPathClean("a/.."), false)
	unittest.Equals(t, IsPathClean("a/./"), false)
	unittest.Equals(t, IsPathClean("a/../"), false)
	unittest.Equals(t, IsPathClean("../a/."), false)
	unittest.Equals(t, IsPathClean("../a/./"), false)
	unittest.Equals(t, IsPathClean("../a/.."), false)
	unittest.Equals(t, IsPathClean("../a/../"), false)
	unittest.Equals(t, IsPathClean("a/../b/c"), false)
	unittest.Equals(t, IsPathClean("a/b/../c"), false)
	unittest.Equals(t, IsPathClean("/."), false)
	unittest.Equals(t, IsPathClean("/.."), false)
	unittest.Equals(t, IsPathClean("/./"), false)
	unittest.Equals(t, IsPathClean("/../"), false)

	// invalid - multiple delimiters
	// filepath.Clean doesn't like multiple /, such as //
	// the reason for this is that this is for paths not URIs
	unittest.Equals(t, IsPathClean(".//"), false)
	unittest.Equals(t, IsPathClean("..//"), false)
	unittest.Equals(t, IsPathClean("..///"), false)
	// this is invalid because // gets cleaned to /
	unittest.Equals(t, IsPathClean("//"), false)
	unittest.Equals(t, IsPathClean("///"), false)

	// valid
	unittest.Equals(t, IsPathClean("/"), true)
	// relative paths are allowed
	unittest.Equals(t, IsPathClean("a"), true)
	unittest.Equals(t, IsPathClean("a/"), true)
	unittest.Equals(t, IsPathClean("a/b"), true)
	unittest.Equals(t, IsPathClean("a/b/"), true)
	unittest.Equals(t, IsPathClean("a/b/c"), true)
	unittest.Equals(t, IsPathClean(".."), true)
	unittest.Equals(t, IsPathClean("../"), true)
	unittest.Equals(t, IsPathClean("../a"), true)
	unittest.Equals(t, IsPathClean("../a/"), true)
}
