package main

import (
	"fmt"
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestFuncs(t *testing.T) {
	log.Println("TestFuncs")

	fmt.Println("IsAppServiceID")
	unittest.Equals(t, IsAppServiceID(""), false)
	unittest.Equals(t, IsAppServiceID("."), false)
	unittest.Equals(t, IsAppServiceID(".."), false)
	unittest.Equals(t, IsAppServiceID("-"), false)
	unittest.Equals(t, IsAppServiceID("a-"), false)
	unittest.Equals(t, IsAppServiceID("-a"), false)
	unittest.Equals(t, IsAppServiceID("a-a"), true)
	unittest.Equals(t, IsAppServiceID("a--a"), true)
	unittest.Equals(t, IsAppServiceID("123456789012345679012345678901234567890123456789012345678901234"), true)
	unittest.Equals(t, IsAppServiceID("1234567890123456790123456789012345678901234567890123456789012345"), false)

	for i := 0; i <= 0x7F; i++ {
		if i >= '0' && i <= '9' ||
			i >= 'A' && i <= 'Z' ||
			i >= 'a' && i <= 'z' {
			unittest.Equals(t, IsAppServiceID(string([]byte{byte(i)})), true)
		} else {
			unittest.Equals(t, IsAppServiceID(string([]byte{byte(i)})), false)
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
