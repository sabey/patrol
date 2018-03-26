package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	ERR_APPS_EMPTY                 = fmt.Errorf("Apps were empty")
	ERR_APP_KEY_EMPTY              = fmt.Errorf("App Key was empty")
	ERR_APP_KEY_INVALID            = fmt.Errorf("App Key was invalid")
	ERR_APP_NAME_EMPTY             = fmt.Errorf("App Name was empty")
	ERR_APP_WORKINGDIRECTORY_EMPTY = fmt.Errorf("App WorkingDirectory was empty")
	ERR_APP_APPPATH_EMPTY          = fmt.Errorf("App App was empty")
	ERR_APP_LOG_DIRECTORY_EMPTY    = fmt.Errorf("App Log Directory was empty")
	ERR_APP_PIDPATH_EMPTY          = fmt.Errorf("App PIDPATH was empty")
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
	Name string `json:"name,omitempty"`
	// if Working Directory is empty, the app is executed from the current working directory
	WorkingDirectory string `json:"working-directory,omitempty"`
	// App Path to the app executable
	AppPath string `json:"app-path,omitempty"`
	// Log Directory for stderr and stdout
	LogDirectory string `json:"log-directory,omitempty"`
	// path to pid file
	PIDPath string `json:"pid-path,omitempty"`
}

func (self *PatrolApp) validate() error {
	if self.Name == "" {
		return ERR_APP_NAME_EMPTY
	}
	if self.WorkingDirectory == "" {
		return ERR_APP_WORKINGDIRECTORY_EMPTY
	}
	if self.AppPath == "" {
		return ERR_APP_APPPATH_EMPTY
	}
	if self.LogDirectory == "" {
		return ERR_APP_LOG_DIRECTORY_EMPTY
	}
	if self.PIDPath == "" {
		return ERR_APP_PIDPATH_EMPTY
	}
	return nil
}
