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
	bs, _ := json.MarshalIndent(config, "", "\t")
	fmt.Printf("config %s\n", bs)
}
