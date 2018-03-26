package main

import (
	"fmt"
)

const (
	// app name maximum length in bytes
	APP_NAME_MAXLENGTH = 255
)

var (
	ERR_APP_NAME_EMPTY                = fmt.Errorf("App Name was empty")
	ERR_APP_NAME_MAXLENGTH            = fmt.Errorf("App Name was longer than 255 bytes")
	ERR_APP_WORKINGDIRECTORY_EMPTY    = fmt.Errorf("App WorkingDirectory was empty")
	ERR_APP_WORKINGDIRECTORY_RELATIVE = fmt.Errorf("App WorkingDirectory was relative")
	ERR_APP_WORKINGDIRECTORY_UNCLEAN  = fmt.Errorf("App WorkingDirectory was unclean")
	ERR_APP_APPPATH_EMPTY             = fmt.Errorf("App AppPath was empty")
	ERR_APP_APPPATH_UNCLEAN           = fmt.Errorf("App AppPath was unclean")
	ERR_APP_LOGDIRECTORY_EMPTY        = fmt.Errorf("App Log Directory was empty")
	ERR_APP_LOGDIRECTORY_UNCLEAN      = fmt.Errorf("App Log Directory was unclean")
	ERR_APP_PIDPATH_EMPTY             = fmt.Errorf("App PIDPATH was empty")
	ERR_APP_PIDPATH_UNCLEAN           = fmt.Errorf("App PIDPath was unclean")
)

type PatrolApp struct {
	Name string `json:"name,omitempty"`
	// Working Directory is currently required to be non empty
	// we don't want Apps executing relative to the current directory, we want them to know what their reference is
	// IF any other path is relative and not absolute, they will be considered relative to the working directory
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
	if len(self.Name) > APP_NAME_MAXLENGTH {
		return ERR_APP_NAME_MAXLENGTH
	}
	if self.WorkingDirectory == "" {
		return ERR_APP_WORKINGDIRECTORY_EMPTY
	}
	if self.WorkingDirectory[0] != '/' {
		// working directory can not be relative
		// we require that it is absolute, so that other paths may be relative to it
		return ERR_APP_WORKINGDIRECTORY_RELATIVE
	}
	if !IsPathClean(self.WorkingDirectory) {
		return ERR_APP_WORKINGDIRECTORY_UNCLEAN
	}
	if self.AppPath == "" {
		return ERR_APP_APPPATH_EMPTY
	}
	if !IsPathClean(self.AppPath) {
		return ERR_APP_APPPATH_UNCLEAN
	}
	if self.LogDirectory == "" {
		return ERR_APP_LOGDIRECTORY_EMPTY
	}
	if !IsPathClean(self.LogDirectory) {
		return ERR_APP_LOGDIRECTORY_UNCLEAN
	}
	if self.PIDPath == "" {
		return ERR_APP_PIDPATH_EMPTY
	}
	if !IsPathClean(self.PIDPath) {
		return ERR_APP_PIDPATH_UNCLEAN
	}
	return nil
}
