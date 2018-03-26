package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	ERR_APPS_EMPTY       = fmt.Errorf("Apps were empty")
	ERR_APPS_KEY_EMPTY   = fmt.Errorf("App Key was empty")
	ERR_APPS_KEY_INVALID = fmt.Errorf("App Key was invalid")
)

func CreatePatrol(
	path string,
) (
	*Patrol,
	error,
) {
	file, err := os.Open(path)
	if err != nil {
		// couldn't open file
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &Patrol{}
	if err := decoder.Decode(config); err != nil {
		// couldn't decode file as json
		return nil, err
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type Patrol struct {
	Apps map[string]*PatrolApp `json:"apps,omitempty"`
}

func (self *Patrol) validate() error {
	if len(self.Apps) == 0 {
		// no apps found
		return ERR_APPS_EMPTY
	}
	for name, app := range self.Apps {
		if name == "" {
			return ERR_APPS_KEY_EMPTY
		}
		if !IsAppKey(name) {
			return ERR_APPS_KEY_INVALID
		}
		if err := app.validate(); err != nil {
			return err
		}
	}
	return nil
}
