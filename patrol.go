package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	ERR_APPS_EMPTY              = fmt.Errorf("Apps were empty")
	ERR_APP_KEY_EMPTY           = fmt.Errorf("App Key was empty")
	ERR_APP_KEY_INVALID         = fmt.Errorf("App Key was invalid")
	ERR_APP_NAME_EMPTY          = fmt.Errorf("App Name was empty")
	ERR_APP_DIRECTORY_EMPTY     = fmt.Errorf("App Directory was empty")
	ERR_APP_FILE_EMPTY          = fmt.Errorf("App File was empty")
	ERR_APP_LOG_DIRECTORY_EMPTY = fmt.Errorf("App Log Directory was empty")
	ERR_APP_PID_EMPTY           = fmt.Errorf("App PID was empty")
)

func CreatePatrol(
	path string,
) (
	*Patrol,
	error,
) {
	file, err := os.Open(path)
	if err != nil {
		//couldn't open file
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &Patrol{}
	if err := decoder.Decode(config); err != nil {
		//"couldn't decode file"
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
		//no apps found
		return ERR_APPS_EMPTY
	}
	for name, app := range self.Apps {
		if name == "" {
			return ERR_APP_KEY_EMPTY
		}
		if !IsAppKey(name) {
			return ERR_APP_KEY_INVALID
		}
		if err := app.validate(); err != nil {
			return err
		}
	}
	return nil
}

type PatrolApp struct {
	Name         string `json:"name,omitempty"`
	Directory    string `json:"directory,omitempty"`
	File         string `json:"file,omitempty"`
	LogDirectory string `json:"log-directory,omitempty"`
	PID          string `json:"pid,omitempty"`
}

func (self *PatrolApp) validate() error {
	if self.Name == "" {
		return ERR_APP_NAME_EMPTY
	}
	if self.Directory == "" {
		return ERR_APP_DIRECTORY_EMPTY
	}
	if self.File == "" {
		return ERR_APP_FILE_EMPTY
	}
	if self.LogDirectory == "" {
		return ERR_APP_LOG_DIRECTORY_EMPTY
	}
	if self.PID == "" {
		return ERR_APP_PID_EMPTY
	}
	return nil
}
