package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
)

var (
	config_path = flag.String("config", "config.json", "path to patrol config file")
)

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	config, err := CreatePatrol(*config_path)
	if err != nil {
		log.Fatalf("failed to create patrol: %s\n", err)
		return
	}
	// when printing a services ignored exit codes, []uint8 will be outputted as a string and it's value is base64 encoded
	// the reason for this is that go may either be assuming this is a string or that it's more efficient to use base64 than to output a bunch of integers
	// except, if we were to use the type int8 or even any other type of int the output would not use base64 as a representation
	// if we were to use the outputted base64 value in our config.json, encoding/json would decode the base64 without problem
	//
	// unmarshalling:
	//	"ignore-exit-codes": "/w=="
	// is the same as:
	//	"ignore-exit-codes": [
	//		255
	//	]
	bs, _ := json.MarshalIndent(config, "", "\t")
	fmt.Printf("config %s\n", bs)
}
