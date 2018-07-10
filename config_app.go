package patrol

import (
	"fmt"
	"io"
	"os"
)

var (
	ERR_APP_NAME_EMPTY                = fmt.Errorf("App Name was empty")
	ERR_APP_NAME_MAXLENGTH            = fmt.Errorf("App Name was longer than 255 bytes")
	ERR_APP_WORKINGDIRECTORY_EMPTY    = fmt.Errorf("App WorkingDirectory was empty")
	ERR_APP_WORKINGDIRECTORY_RELATIVE = fmt.Errorf("App WorkingDirectory was relative")
	ERR_APP_WORKINGDIRECTORY_UNCLEAN  = fmt.Errorf("App WorkingDirectory was unclean")
	ERR_APP_BINARY_EMPTY              = fmt.Errorf("App Binary was empty")
	ERR_APP_BINARY_UNCLEAN            = fmt.Errorf("App Binary was unclean")
	ERR_APP_LOGDIRECTORY_EMPTY        = fmt.Errorf("App Log Directory was empty")
	ERR_APP_LOGDIRECTORY_UNCLEAN      = fmt.Errorf("App Log Directory was unclean")
	ERR_APP_KEEPALIVE_INVALID         = fmt.Errorf("App KeepAlive was invalid, please select a keep alive method!")
	ERR_APP_PIDPATH_EMPTY             = fmt.Errorf("App PIDPATH was empty")
	ERR_APP_PIDPATH_UNCLEAN           = fmt.Errorf("App PIDPath was unclean")
	ERR_APP_EXECUTETIMEOUT_INVALID    = fmt.Errorf("App Excute Timeout < 0")
)

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
	// if Disabled is true the App won't be executed until enabled
	// the only way to enable this once loaded is to use an API or restart Patrol
	// if Disabled is true the App MAY be running, we will just avoid watching it!
	Disabled bool `json:"disabled,omitempty"`
	// clear keyvalue on new instance?
	KeyValueClear bool `json:"keyvalue-clear,omitempty"`
	////////////
	// os.Cmd //
	////////////
	// we can optionally run our executable for a set duration
	// timeout is in Seconds
	// 0 is disabled by default
	//
	// it's unclear what will happen with this if our process exits
	// we may want to enable this only for APP_KEEPALIVE_PID_PATROL - because thats the only time we will have full control over a process
	ExecuteTimeout int `json:"execute-timeout,omitempty"`
	// Args holds command line arguments, including the command as Args[0].
	// If the Args field is empty or nil, Run uses {Path}.
	//
	// In typical use, both Path and Args are set by calling Command.
	Args []string `json:"args,omitempty"`
	// os.Cmd.Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	// If os.Cmd.Env is nil, the new process uses the current process's environment.
	// If os.Cmd.Env contains duplicate environment keys, only the last value in the slice for each duplicate key is used.
	//
	// we're going to include our own patrol related environment variables
	// so EnvParent would be important, since if a user expects nil Env they'll never get parent variables
	Env []string `json:"env,omitempty"`
	// if we want to optionally include our own args AND and still keep our original parent args we MUST use this!
	EnvParent bool `json:"env-parent,omitempty"`
	// these are NOT supported with JSON for obvious reasons
	// these will have to be set manually!!!
	// extra execute options
	// these can dynamically return new values on each execute
	//
	// extra args are appended at the very end, so they could overwrite anything that comes before
	ExtraArgs func(
		app *App,
	) []string `json:"-"`
	// extra env are appended at the very end, so they could overwrite anything that comes before
	ExtraEnv func(
		app *App,
	) []string `json:"-"`
	// Stdin, Stdout, Stderr// Stdin specifies the process's standard input.
	//
	// If Stdin is nil, the process reads from the null device (os.DevNull).
	//
	// If Stdin is an *os.File, the process's standard input is connected
	// directly to that file.
	//
	// Otherwise, during the execution of the command a separate
	// goroutine reads from Stdin and delivers that data to the command
	// over a pipe. In this case, Wait does not complete until the goroutine
	// stops copying, either because it has reached the end of Stdin
	// (EOF or a read error) or because writing to the pipe returned an error.
	Stdin io.Reader `json:"-"`
	// Stdout and Stderr specify the process's standard output and error.
	//
	// If either is nil, Run connects the corresponding file descriptor
	// to the null device (os.DevNull).
	//
	// If either is an *os.File, the corresponding output from the process
	// is connected directly to that file.
	//
	// Otherwise, during the execution of the command a separate goroutine
	// reads from the process over a pipe and delivers that data to the
	// corresponding Writer. In this case, Wait does not complete until the
	// goroutine reaches EOF or encounters an error.
	//
	// If Stdout and Stderr are the same writer, and have a type that can
	// be compared with ==, at most one goroutine at a time will call Write.
	Stdout io.Writer `json:"-"`
	Stderr io.Writer `json:"-"`
	// ExtraFiles specifies additional open files to be inherited by the
	// new process. It does not include standard input, standard output, or
	// standard error. If non-nil, entry i becomes file descriptor 3+i.
	//
	// this is not a slice since we may want to change it in the future
	// std err/out/err is easier to deal with since you can wrap it anyway you want
	ExtraFiles func(
		app *App,
	) []*os.File `json:"-"`
	// Triggers
	// we're going to allow our Start function to overwrite if our App is Disabled
	// this will be just incase we want to hold a disabled state outside of this app, such as in a database, just incase we crash
	// we'll check the value of App.disabled on return
	TriggerStart func(
		app *App,
	) `json:"-"`
	TriggerStarted func(
		app *App,
	) `json:"-"`
	TriggerStartedPinged func(
		app *App,
	) `json:"-"`
	TriggerStartFailed func(
		app *App,
	) `json:"-"`
	TriggerRunning func(
		app *App,
	) `json:"-"`
	TriggerClosed func(
		app *App,
		history *History,
	) `json:"-"`
	TriggerPinged func(
		app *App,
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
	o := &ConfigApp{
		KeepAlive:            self.KeepAlive,
		Name:                 self.Name,
		Binary:               self.Binary,
		WorkingDirectory:     self.WorkingDirectory,
		LogDirectory:         self.LogDirectory,
		PIDPath:              self.PIDPath,
		PIDVerify:            self.PIDVerify,
		Disabled:             self.Disabled,
		KeyValueClear:        self.KeyValueClear,
		ExecuteTimeout:       self.ExecuteTimeout,
		Args:                 make([]string, 0, len(self.Args)),
		Env:                  make([]string, 0, len(self.Env)),
		EnvParent:            self.EnvParent,
		ExtraArgs:            self.ExtraArgs,
		ExtraEnv:             self.ExtraEnv,
		Stdin:                self.Stdin,
		Stdout:               self.Stdout,
		Stderr:               self.Stderr,
		ExtraFiles:           self.ExtraFiles,
		TriggerStart:         self.TriggerStart,
		TriggerStarted:       self.TriggerStarted,
		TriggerStartedPinged: self.TriggerStartedPinged,
		TriggerStartFailed:   self.TriggerStartFailed,
		TriggerRunning:       self.TriggerRunning,
		TriggerClosed:        self.TriggerClosed,
		TriggerPinged:        self.TriggerPinged,
	}
	for _, a := range self.Args {
		o.Args = append(o.Args, a)
	}
	for _, e := range self.Env {
		o.Env = append(o.Env, e)
	}
	return o
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
	if self.ExecuteTimeout < 0 {
		return ERR_APP_EXECUTETIMEOUT_INVALID
	}
	return nil
}
