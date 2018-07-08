package patrol

type ConfigApp struct {
	// keep alive method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your app may function
	KeepAlive int `json:"keepalive,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
	Name string `json:"name,omitempty"`
	// Binary is the path to the executable
	Binary string `json:"binary,omitempty"`
	// Working Directory is currently required to be non empty
	// we don't want Apps executing relative to the current directory, we want them to know what their reference is
	// IF any other path is relative and not absolute, they will be considered relative to the working directory
	WorkingDirectory string `json:"working-directory,omitempty"`
	// Log Directory for stderr and stdout
	LogDirectory string `json:"log-directory,omitempty"`
	// path to pid file
	// PID is optional, it is only required when using the PATROL or APP keepalive methods
	PIDPath string `json:"pid-path,omitempty"`
	// should we verify that the PID belongs to Binary?
	// the reason for this is that it is technically possible for your App to write a PID to file, exit, and then for another long running service to start with this same PID
	// the problem here is that that other long running process would be confused for our App and we would assume it is running
	// the only solution is to verify the processes name OR for you to continuously write your PID to file in intervals, say write to PID every 10 seconds
	// the problem with the constant PID writing is that should your parent fork and create a child, you would want to stop writing the parent PID and only write the child PID!
	PIDVerify bool `json:"pid-verify,omitempty"`
	// these are NOT supported with JSON for obvious reasons
	// these will have to be set manually!!!
	// Triggers
	TriggerStart func(
		id string,
		app *App,
	) `json:"-"`
	TriggerStarted func(
		id string,
		app *App,
	) `json:"-"`
	TriggerStartFailed func(
		id string,
		app *App,
	) `json:"-"`
	TriggerRunning func(
		id string,
		app *App,
	) `json:"-"`
	TriggerStopped func(
		id string,
		app *App,
		history *History,
	) `json:"-"`
}

func (self *ConfigApp) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *ConfigApp) Clone() *ConfigApp {
	if self == nil {
		return nil
	}
	return &ConfigApp{
		KeepAlive:          self.KeepAlive,
		Name:               self.Name,
		Binary:             self.Binary,
		WorkingDirectory:   self.WorkingDirectory,
		LogDirectory:       self.LogDirectory,
		PIDPath:            self.PIDPath,
		PIDVerify:          self.PIDVerify,
		TriggerStart:       self.TriggerStart,
		TriggerStarted:     self.TriggerStarted,
		TriggerStartFailed: self.TriggerStartFailed,
		TriggerRunning:     self.TriggerRunning,
		TriggerStopped:     self.TriggerStopped,
	}
}
func (self *ConfigApp) Validate() error {
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
	if self.Binary == "" {
		return ERR_APP_BINARY_EMPTY
	}
	if !IsPathClean(self.Binary) {
		return ERR_APP_BINARY_UNCLEAN
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
