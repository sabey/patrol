package main

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestFuncs(t *testing.T) {
	log.Println("TestFuncs")

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
}
