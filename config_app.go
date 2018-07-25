package patrol

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
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
	// KeepAlive Method
	//
	// APP_KEEPALIVE_PID_PATROL = 1
	// APP_KEEPALIVE_PID_APP = 2
	// APP_KEEPALIVE_HTTP = 3
	// APP_KEEPALIVE_UDP = 4
	//
	// PID_PATROL: Patrol will watch the execution of the Application. Apps will not be able to fork.
	// PID_APP: The Application is required to write its CURRENT PID to our `pid-path`. Patrol will `kill -0 PID` to verify that the App is running. This option should be used for forking processes.
	// HTTP: The Application must send a Ping request to our HTTP API.
	// UDP: The Application must send a Ping request to our UDP API.
	KeepAlive int `json:"keepalive,omitempty"`
	// Name is used as our Display Name in our HTTP GUI.
	// Name can contain any characters but must be less than 255 bytes in length.
	Name string `json:"name,omitempty"`
	// Binary is the relative path to the executable
	Binary string `json:"binary,omitempty"`
	// Working Directory is the ABSOLUTE Path to our Application Directory.
	// Binary, LogDirectory, and PidPath are RELATIVE to this Path!
	//
	// The only time WorkingDirectory is allowed to be relative is if we're prefixed with ~/
	// If prefixed with ~/, we will then replace it with our current users home directory.
	WorkingDirectory string `json:"working-directory,omitempty"`
	// Log Directory is the relative path to our log directory.
	// STDErr and STDOut Logs are held in a `YEAR/MONTH/DAY` sub folder.
	LogDirectory string `json:"log-directory,omitempty"`
	// Path is the relative path to our PID file.
	// PID is optional, it is only required when using the KeepAlive method: APP_KEEPALIVE_PID_APP
	// Our PID file must ONLY contain the integer of our current PID
	PIDPath string `json:"pid-path,omitempty"`
	// PIDVerify - Should we verify that our PID belongs to Binary?
	// PIDVerify is optional, it is only supported when using the KeepAlive method: APP_KEEPALIVE_PID_APP
	// This currently is NOT supported.
	// By default when we execute an App - `ps aux` will report our FULL PATH and BINARY as our first Arg.
	// If our process should fork, we're unsure of how this will change. We may have to compare that PID contains at the very least Binary in the first Arg.
	PIDVerify bool `json:"pid-verify,omitempty"`
	// If Disabled is true our App won't be executed until enabled.
	// The only way to enable an App once Patrol is started is to use the API or restart Patrol
	// If we are Disabled and we discover an App that is running, we will signal it to stop.
	Disabled bool `json:"disabled,omitempty"`
	// KeyValue - prexisting values to populate objects with on init
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
	// KeyValueClear if true will cause our App KeyValue to be cleared once a new instance of our App is started.
	KeyValueClear bool `json:"keyvalue-clear,omitempty"`
	// If Secret is set, we will require a secret to be passed when pinging and modifying the state of our App from our HTTP and UDP API.
	// We are not going to throttle comparing our secret. Choose a secret with enough bits of uniqueness and don't make your Patrol instance public!
	// If you are worried about your secret being public, use TLS and HTTP, DO NOT USE UDP!!!
	Secret string `json:"secret,omitempty"`
	////////////
	// os.Cmd //
	////////////
	// ExecuteTimeout is an optional value in seconds of how long we will run our App for.
	// A Value of 0 will disable this.
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
	// We're going to include our own Patrol related environment variables, so EnvParent is required if we wish to include parent values.
	Env []string `json:"env,omitempty"`
	// If EnvParent is true, we will prepend all of our Patrol environment variables to the execution of our process.
	EnvParent bool `json:"env-parent,omitempty"`
	// These options are only available when you extend Patrol as a library
	// These values will NOT be able to be set from `config.json` - They must be set manually
	//
	// ExtraArgs is an optional set of values that will be appended to Args.
	ExtraArgs func(
		app *App,
	) []string `json:"-"`
	// ExtraEnv is an optional set of values that will be appended to Env.
	ExtraEnv func(
		app *App,
	) []string `json:"-"`
	// Stdin specifies the process's standard input.
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
	//
	// If this value is nil, Patrol will create its own file located in our Log Directory.
	// If this value is nil, this file will also be able to be read from the HTTP GUI.
	Stdout io.Writer `json:"-"`
	Stderr io.Writer `json:"-"`
	// Merge Stdout and Stderr into a single file?
	StdMerge bool `json:"std-merge,omitempty"`
	// ExtraFiles specifies additional open files to be inherited by the
	// new process. It does not include standard input, standard output, or
	// standard error. If non-nil, entry i becomes file descriptor 3+i.
	ExtraFiles func(
		app *App,
	) []*os.File `json:"-"`
	// TriggerStart is called from tick in runApps() before we attempt to execute an App.
	TriggerStart func(
		app *App,
	) `json:"-"`
	// TriggerStarted is called from tick in runApps() and isAppRunning()
	// This is called after we either execute a new App or we discover a newly running App.
	TriggerStarted func(
		app *App,
	) `json:"-"`
	// TriggerStartedPinged is called from App.apiRequest() when we discover a newly running App from a Ping request.
	TriggerStartedPinged func(
		app *App,
	) `json:"-"`
	// TriggerStartFailed is called from tick in runApps() when we fail to execute a new App.
	TriggerStartFailed func(
		app *App,
	) `json:"-"`
	// TriggerRunning is called from tick() when we discover an App is running.
	TriggerRunning func(
		app *App,
	) `json:"-"`
	// TriggerDisabled is called from tick() when we discover an App that is disabled.
	TriggerDisabled func(
		app *App,
	) `json:"-"`
	// TriggerClosed is called from App.close() when we discover a previous instance of an App is closed.
	TriggerClosed func(
		app *App,
		history *History,
	) `json:"-"`
	// TriggerPinged is from App.apiRequest() when we discover an App is running from a Ping request.
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
		KeyValue:             make(map[string]interface{}),
		KeyValueClear:        self.KeyValueClear,
		Secret:               self.Secret,
		ExecuteTimeout:       self.ExecuteTimeout,
		Args:                 make([]string, 0, len(self.Args)),
		Env:                  make([]string, 0, len(self.Env)),
		EnvParent:            self.EnvParent,
		ExtraArgs:            self.ExtraArgs,
		ExtraEnv:             self.ExtraEnv,
		Stdin:                self.Stdin,
		Stdout:               self.Stdout,
		Stderr:               self.Stderr,
		StdMerge:             self.StdMerge,
		ExtraFiles:           self.ExtraFiles,
		TriggerStart:         self.TriggerStart,
		TriggerStarted:       self.TriggerStarted,
		TriggerStartedPinged: self.TriggerStartedPinged,
		TriggerStartFailed:   self.TriggerStartFailed,
		TriggerRunning:       self.TriggerRunning,
		TriggerDisabled:      self.TriggerDisabled,
		TriggerClosed:        self.TriggerClosed,
		TriggerPinged:        self.TriggerPinged,
	}
	for k, v := range self.KeyValue {
		o.KeyValue[k] = v
	}
	for _, a := range self.Args {
		o.Args = append(o.Args, a)
	}
	for _, e := range self.Env {
		o.Env = append(o.Env, e)
	}
	// fix WorkingDirectory
	if strings.HasPrefix(o.WorkingDirectory, "~/") {
		// remove ~/
		// prefix with Home Directory
		// if we can't get a homedir we're not going to modify our working directory
		// if we fail to fix this path we will get an error saying the working directory is relative
		u, err := user.Current()
		if err != nil {
			log.Printf("./patrol.ConfigApp.Clone(): failed to get user.Current(): \"%s\"\n", err)
		} else {
			// we need homedir to be non empty before we trim
			// we don't want to make our relative path absolute if it is empty
			if u.HomeDir != "" {
				// just trim ~
				o.WorkingDirectory = fmt.Sprintf("%s%s", u.HomeDir, o.WorkingDirectory[1:])
			}
		}
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
	if len(self.Secret) > SECRET_MAX_LENGTH {
		return ERR_SECRET_TOOLONG
	}
	if self.ExecuteTimeout < 0 {
		return ERR_APP_EXECUTETIMEOUT_INVALID
	}
	return nil
}
