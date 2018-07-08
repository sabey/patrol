package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	ERR_CONFIG_NIL           = fmt.Errorf("Config was NIL")
	ERR_PATROL_EMPTY         = fmt.Errorf("Patrol Apps and Servers were both empty")
	ERR_APPS_KEY_EMPTY       = fmt.Errorf("App Key was empty")
	ERR_APPS_KEY_INVALID     = fmt.Errorf("App Key was invalid")
	ERR_APPS_APP_NIL         = fmt.Errorf("App was nil")
	ERR_SERVICES_KEY_EMPTY   = fmt.Errorf("Service Key was empty")
	ERR_SERVICES_KEY_INVALID = fmt.Errorf("Service Key was invalid")
	ERR_SERVICES_SERVICE_NIL = fmt.Errorf("Service was nil")
)
var (
	// do not change this in this package
	// change it in testing files or for debugging purposes only
	unittesting bool = false
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
	if config == nil {
		// decoded config was nil
		// while this would rarely occur in practice, it is technically possible for config to be nil after decode/unmarshal
		// a json file with the value of "null" would cause this to occur
		// the only time this really plays out in reality is if someone were to POST null to a JSON API
		// this something that is sometimes overlooked
		return nil, ERR_CONFIG_NIL
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type Patrol struct {
	// Apps must contain a unique non empty key: ( 0-9 A-Z a-z - . ) and must not be "." or ".."
	// this key is used for the HTTP/UDP endpoints and is used to represent Apps in our log files, PatrolApp.Name is not used in logs
	Apps map[string]*PatrolApp `json:"apps,omitempty"`
	// Services must contain a unique non empty key: ( 0-9 A-Z a-z - . ) and must not be "." or ".."
	// this key is used for inplace of the "service * status" or "/etc/init.d/* status" field
	Services map[string]*PatrolService `json:"services,omitempty"`
}

func (self *Patrol) validate() error {
	if len(self.Apps) == 0 &&
		len(self.Services) == 0 {
		// no apps or services found
		return ERR_PATROL_EMPTY
	}
	// check apps
	for name, app := range self.Apps {
		if name == "" {
			return ERR_APPS_KEY_EMPTY
		}
		if !IsAppServiceKey(name) {
			return ERR_APPS_KEY_INVALID
		}
		if !app.IsValid() {
			return ERR_APPS_APP_NIL
		}
		if err := app.validate(); err != nil {
			return err
		}
	}
	// check services
	for name, service := range self.Services {
		if name == "" {
			return ERR_SERVICES_KEY_EMPTY
		}
		if !IsAppServiceKey(name) {
			return ERR_SERVICES_KEY_INVALID
		}
		if !service.IsValid() {
			return ERR_SERVICES_SERVICE_NIL
		}
		if err := service.validate(); err != nil {
			return err
		}
	}
	return nil
}
