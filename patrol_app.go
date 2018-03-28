package main

import (
	"fmt"
)

const (
	// app name maximum length in bytes
	APP_NAME_MAXLENGTH = 255
)

// there are multiple methods of process management, none of them are perfect! they all have their tradeoffs!!!
const (
	// KeepAlive is controlled by Patrol:
	// after executing App, Patrol will write the PID to file
	// the upside to this is that we can know exactly when our App exits - we won't have to constantly send "kill -0 PID" to our App to check if it's still alive
	// the downside to this is that if App were to fork and the parent process were to exit we would not be able to track the forked processes, we would then have to respawn our App
	// we're trading fast respawns and external PID access for no child forking
	APP_KEEPALIVE_PID_PATROL = iota + 1
	// KeepAlive is controlled by App:
	// once an App spawns the App is required to write its own PID to file
	// the upside to this is that if the parent process were to fork, the child process would write its new PID to file and Patrol would read and monitor that latest PID instead
	// the downside to this is that we constantly have to load the PID from file and send "kill -0 PID" to check if the process is still alive
	// the other downside is that we may want to optionally check that the PID belongs to AppPath, and if it were not we would have to respawn our App
	// this will allow easy child forking and the ability for the parent process to exit after forking
	// the other trade off is that we won't be able to see the exact time when our monitored PID process exits, leaving a possible delay between respawn
	// see further notes at PatrolApp.PIDVerify
	APP_KEEPALIVE_PID_APP
	// KeepAlive is controlled by HTTP:
	// once an App spawns the App is required intermittently send a HTTP POST to the Patrol JSON API as the keepalive method
	// the upside to this is that we don't have to monitor a PID file and it supports child forking
	// while we won't monitor the PID file here the HTTP POST is required to POST a PID, this way we can send a signal to the PID from the gui admin page for example
	// the downside is that the App must support HTTP clients and have a way to easily do intermittent HTTP POSTs, this might be a burden for most Apps
	APP_KEEPALIVE_HTTP
	// KeepAlive is controlled by UDP:
	// this is similar to HTTP except that it requires less overhead with the added downside that your App does not get a returned confirmation that Patrol is receiving your pings
	// this could be a steep price to pay should the Patrol UDP listener become unresponsive or your UDP packets never arrive, this would result in the App respawning
	APP_KEEPALIVE_UDP
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
	ERR_APP_KEEPALIVE_INVALID         = fmt.Errorf("App KeepAlive was invalid, please select a keep alive method!")
	ERR_APP_PIDPATH_EMPTY             = fmt.Errorf("App PIDPATH was empty")
	ERR_APP_PIDPATH_UNCLEAN           = fmt.Errorf("App PIDPath was unclean")
)

type PatrolApp struct {
	// keep alive method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your app may function
	KeepAlive int `json:"keepalive,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
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
	// PID is optional, it is only required when using the PATROL or APP keepalive methods
	PIDPath string `json:"pid-path,omitempty"`
	// should we verify that the PID belongs to AppPath?
	// the reason for this is that it is technically possible for your App to write a PID to file, exit, and then for another long running service to start with this same PID
	// the problem here is that that other long running process would be confused for our App and we would assume it is running
	// the only solution is to verify the processes name OR for you to continuously write your PID to file in intervals, say write to PID every 10 seconds
	// the problem with the constant PID writing is that should your parent fork and create a child, you would want to stop writing the parent PID and only write the child PID!
	PIDVerify bool `json:"pid-verify,omitempty"`
}

func (self *PatrolApp) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *PatrolApp) validate() error {
	if self.KeepAlive < APP_KEEPALIVE_PID_PATROL ||
		self.KeepAlive > APP_KEEPALIVE_UDP {
		// unknown keep alive value
		return ERR_APP_KEEPALIVE_INVALID
	}
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
	if self.KeepAlive == APP_KEEPALIVE_PID_PATROL ||
		self.KeepAlive == APP_KEEPALIVE_PID_APP {
		// PID is required
		if self.PIDPath == "" {
			return ERR_APP_PIDPATH_EMPTY
		}
		if !IsPathClean(self.PIDPath) {
			return ERR_APP_PIDPATH_UNCLEAN
		}
	}
	return nil
}
