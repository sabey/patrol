# Patrol
### Process and Service Management
[![GoDoc](https://godoc.org/sabey.co/patrol?status.svg)](https://godoc.org/sabey.co/patrol)
![Patrol GUI](https://sabey.co/patrol-gui.png)


## Building Patrol
```bash
go get -t -u sabey.co/patrol
cd ~/go/src/sabey.co/patrol/patrol
go build -a -v
./patrol
# HTML GUI: http://localhost:8421
```


## [Installing Patrol - systemd](https://github.com/sabey/patrol/tree/master/patrol)
[patrol/README.md](https://github.com/sabey/patrol/blob/master/patrol/README.md)


#### Patrol config.json
```json
{
  "apps": {
    "testapp": {
      "keepalive": 1,
      "name": "Test App",
      "binary": "testapp",
      "working-directory": "~/go/src/sabey.co/patrol/unittest/testapp/",
      "log-directory": "logs",
      "pid-path": "app.pid",
      "pid-verify": false,
      "disabled": false,
      "keyvalue-clear": true,
      "secret": "",
      "execute-timeout": 0,
      "args": [
        "-config",
        "\"config.json\""
      ],
      "env": [
        "a=b",
        "c=d"
      ],
      "env-parent": false,
      "std-merge": true
    }
  },
  "services": {
    "ssh": {
      "management": 1,
      "name": "SSH",
      "service": "ssh",
      "ignore-exit-codes": [
        127
      ],
      "disabled": false,
      "keyvalue-clear": true,
      "secret": ""
    }
  }
}
```


#### Patrol App Environment Variables
```bash
PATROL_ID=testapp
PATROL_KEEPALIVE=1
PATROL_PID=/path/to/app.pid
PATROL_HTTP=["127.0.0.1:8421"]
PATROL_UDP=["127.0.0.1:1248"]
```


## Unit Tests
```bash
cd ~/go/src/sabey.co/patrol/unittest/testapp
go build -a -v
cd ~/go/src/sabey.co/patrol/unittest/testserver
go build -a -v
cd ~/go/src/sabey.co/patrol/
clear && go vet && go test -race
```


## API

#### HTTP API Endpoint
```bash
GET /status/
# returns API_Status Object

GET /api/?group=(app||service)&id=testapp&toggle=STATE&history=true&secret=SECRET&cas=CAS
# returns API_Response Object

POST /api/
# requires API_Request Object
# returns API_Response Object

```

#### UDP API Endpoint
```bash
127.0.0.1:1248
# requires API_Request Object
# returns API_Response Object to the dialing IP address if no error occurred

```

#### Example API_Request
```json
{
  "id": "http",
  "group": "app",
  "ping": true,
  "pid": 10732
}
```

#### Example API_Response
```json
{
  "id": "http",
  "instance-id": "fd1b743c-752d-4034-a502-2bdddb7e262f",
  "group": "app",
  "name": "testapp",
  "pid": 10732,
  "started": "Wed, 18 Jul 2018 18:43:44 -0700",
  "lastseen": "Wed, 18 Jul 2018 18:43:47 -0700",
  "cas": 57684
}
```

#### Example API_Status
```json
{
  "instance-id": "5127ce9b-61e3-4818-a0fd-f1bebafb2fac",
  "apps": {
    "fake-secret": {
      "name": "Fake Secret App",
      "disabled": true,
      "secret": true,
      "cas": 40875
    },
    "testapp": {
      "instance-id": "08b63d19-7569-4b2d-9ad1-87c780ea2f83",
      "name": "Test App",
      "pid": 26200,
      "started": "Wed, 18 Jul 2018 16:20:14 -0700",
      "lastseen": "Wed, 18 Jul 2018 16:20:48 -0700",
      "cas": 75662
    }
  },
  "service": {
    "ssh": {
      "instance-id": "29a6b077-6af3-4865-b443-4676b242588d",
      "name": "SSH",
      "started": "Wed, 18 Jul 2018 16:20:14 -0700",
      "lastseen": "Wed, 18 Jul 2018 16:20:44 -0700",
      "cas": 75180
    }
  },
  "started": "Wed, 18 Jul 2018 16:20:14 -0700"
}
```


## type Config struct {
```golang
// Apps/Services must contain a unique non empty key: ( 0-9 A-Z a-z - )
// ID MUST be usable as a valid hostname label, ie: len <= 63 AND no starting/ending -
// Keys are NOT our binary name
// Keys are only used as unique identifiers for our API and Keep Alive
Apps     map[string]*ConfigApp     `json:"apps,omitempty"`
Services map[string]*ConfigService `json:"services,omitempty"`

// TickEvery is an integer value in seconds of how often we will check the state of our Apps and Services
// Value of 0 Defaults to 15 seconds
TickEvery int `json:"tick-every,omitempty"`

// History is the maximum amount of instance history we should hold
// Value of 0 Defaults to 100
History int `json:"history,omitempty"`

// Timestamp Layout is used by the JSON API and HTTP GUI templates
//
// Timestamp Layout can be found here:
// https://golang.org/pkg/time/#pkg-constants
// https://golang.org/pkg/time/#example_Time_Format
//
// The recommended value is RFC1123Z: "Mon, 02 Jan 2006 15:04:05 -0700"
//
// An empty value will default to time.String()
// https://golang.org/pkg/time/#Time.String
// This default is: "2006-01-02 15:04:05.999999999 -0700 MST"
// This default will also include our monotonic clock as a suffix: "m=±<value>"
Timestamp string `json:"json-timestamp,omitempty"`

// PingTimeout is an integer value in seconds of how often we require a Ping to be sent
// This only applies to App KeepAlives: APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
PingTimeout int `json:"ping-timeout,omitempty"`

// ListenHTTP/ListenUDP is our list of listeners
// These values are passed as Environment Variables to our executed Apps as JSON Arrays
//
// Example Environment Variables:
// PATROL_HTTP=["127.0.0.1:8421"]
// PATROL_UDP=["127.0.0.1:1248"]
//
// When using APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP, these are the addresses we MUST ping
ListenHTTP []string `json:"listen-http,omitempty"`
ListenUDP  []string `json:"listen-udp,omitempty"`

// HTTP/UDP currently only support the attribute `listen`
// This will allow us to overwrite our default listeners for HTTP and UDP
// In the future this will include additional options.
HTTP *ConfigHTTP `json:"http,omitempty"`
UDP  *ConfigUDP  `json:"udp,omitempty"`


// Triggers are only available when you extend Patrol as a library
// These values will NOT be able to be set from `config.json` - They must be set manually

// TriggerStart is called on CreatePatrol
// This will only be called ONCE
// If an error is returned a Patrol object will NOT be returned!
TriggerStart func(
	patrol *Patrol,
) error `json:"-"`

// TriggerShutdown is called when we call Patrol.Shutdown()
// This will only be called ONCE
// Once Patrol.Shutdown() is called our Patrol object will no longer be usable
TriggerShutdown func(
	patrol *Patrol,
) `json:"-"`

// TriggerStarted is called every time we call Patrol.Start()
TriggerStarted func(
	patrol *Patrol,
) `json:"-"`

// TriggerTick is called every time we Patrol.tick() and BEFORE we check our App and Service States
TriggerTick func(
	patrol *Patrol,
) `json:"-"`

// TriggerStopped is called every time we call Patrol.Stop()
TriggerStopped func(
	patrol *Patrol,
) `json:"-"`

// Extra Unstructured Data
X json.RawMessage `json:"x,omitempty"`
```


## type ConfigApp struct {
```golang
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

// ExtraArgs is an optional set of values that will be appended to Args.
ExtraArgs func(
	id string,
) []string `json:"-"`

// ExtraEnv is an optional set of values that will be appended to Env.
ExtraEnv func(
	id string,
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
	id string,
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

// TriggerShutdown is called when we call Patrol.Shutdown()
// This will only be called ONCE
// This is called regardless if our App is running or disabled!
TriggerShutdown func(
  app *App,
) `json:"-"`

// Extra Unstructured Data
X json.RawMessage `json:"x,omitempty"`
```


## type ConfigService struct {
```golang
// Management Method
//
// SERVICE_MANAGEMENT_SERVICE = 1
// SERVICE_MANAGEMENT_INITD = 2
//
// SERVICE_MANAGEMENT_SERVICE: Patrol will use the command `service *`
// SERVICE_MANAGEMENT_INITD: Patrol will use the command `/etc/init.d/*`
//
// If Management is set it will ignore all of the Management Start/Status/Stop/Restart values
// If Management is 0, Start/Status/Stop/Restart must each be individually set!
// If for whatever reason is necessary, we could choose to user `service` for `status` and `/etc/init.d/` for start or stop!
Management        int `json:"management,omitempty"`
ManagementStart   int `json:"management-start,omitempty"`
ManagementStatus  int `json:"management-status,omitempty"`
ManagementStop    int `json:"management-stop,omitempty"`
ManagementRestart int `json:"management-restart,omitempty"`

// Optionally we may override our service parameters.
// For example, instead of `restart` we may choose to use `force-reload`
ManagementStartParameter   string `json:"management-start-parameter,omitempty"`
ManagementStatusParameter  string `json:"management-status-parameter,omitempty"`
ManagementStopParameter    string `json:"management-stop-parameter,omitempty"`
ManagementRestartParameter string `json:"management-restart-parameter,omitempty"`

// Name is used as our Display Name in our HTTP GUI.
// Name can contain any characters but must be less than 255 bytes in length.
Name string `json:"name,omitempty"`

// Service is the parameter of our service.
// This is the equivalent of Binary
Service string `json:"service,omitempty"`

// These are a list of valid exit codes to ignore when returned from Start/Status/Stop/Restart
// By Default 0 is always ignored, it is assumed to mean that the command was successful!
IgnoreExitCodesStart   []uint8 `json:"ignore-exit-codes-start,omitempty"`
IgnoreExitCodesStatus  []uint8 `json:"ignore-exit-codes-status,omitempty"`
IgnoreExitCodesStop    []uint8 `json:"ignore-exit-codes-stop,omitempty"`
IgnoreExitCodesRestart []uint8 `json:"ignore-exit-codes-restart,omitempty"`

// If Disabled is true our Service won't be executed until enabled.
// The only way to enable an Service once Patrol is started is to use the API or restart Patrol
// If we are Disabled and we discover an Service that is running, we will signal it to stop.
Disabled bool `json:"disabled,omitempty"`

// KeyValue - prexisting values to populate objects with on init
KeyValue map[string]interface{} `json:"keyvalue,omitempty"`

// KeyValueClear if true will cause our Service KeyValue to be cleared once a new instance of our Service is started.
KeyValueClear bool `json:"keyvalue-clear,omitempty"`

// If Secret is set, we will require a secret to be passed when pinging and modifying the state of our Service from our HTTP and UDP API.
// We are not going to throttle comparing our secret. Choose a secret with enough bits of uniqueness and don't make your Patrol instance public!
// If you are worried about your secret being public, use TLS and HTTP, DO NOT USE UDP!!!
Secret string `json:"secret,omitempty"`


// Triggers are only available when you extend Patrol as a library
// These values will NOT be able to be set from `config.json` - They must be set manually

// TriggerStart is called from tick in runServices() before we attempt to execute an Service.
TriggerStart func(
	service *Service,
) `json:"-"`

// TriggerStarted is called from tick in runServices() and isServiceRunning()
// This is called after we either execute a new Service or we discover a newly running Service.
TriggerStarted func(
	service *Service,
) `json:"-"`

// TriggerStartFailed is called from tick in runServices() when we fail to execute a new Service.
TriggerStartFailed func(
	service *Service,
) `json:"-"`

// TriggerRunning is called from tick() when we discover an Service is running.
TriggerRunning func(
	service *Service,
) `json:"-"`

// TriggerDisabled is called from tick() when we discover an Service that is disabled.
TriggerDisabled func(
	service *Service,
) `json:"-"`

// TriggerPinged is from Service.apiRequest() when we discover an Service is running from a Ping request.
TriggerClosed func(
  service *Service,
  history *History,
) `json:"-"`

// TriggerShutdown is called when we call Patrol.Shutdown()
// This will only be called ONCE
// This is called regardless if our Service is running or disabled!
TriggerShutdown func(
  service *Service,
) `json:"-"`

// Extra Unstructured Data
X json.RawMessage `json:"x,omitempty"`
```


## type API_Status struct {
```golang
// Instance ID - UUIDv4
InstanceID string                   `json:"instance-id,omitempty"`
Apps       map[string]*API_Response `json:"apps,omitempty"`
Services   map[string]*API_Response `json:"service,omitempty"`

// Timestamp Patrol started at
Started string `json:"started,omitempty"`

// Is Patrol in a Shutdown state?
Shutdown bool `json:"shutdown,omitempty"`
```


## type API_Request struct {
```golang
// Requests by Default are STATELESS - If no values are set then nothing is modified!
// The reason we're stateless by default is so that our UDP endpoint can make requests as if it were a HTTP GET Request
// UDP has the downside that if an error occurs a response will not be sent in return

// Unique Identifier
ID string `json:"id,omitempty"`

// Group: `app` or `service`
Group string `json:"group,omitempty"`

// Ping?
// Only supported by either APP_KEEPALIVE_HTTP or APP_KEEPALIVE_UDP
// If APP_KEEPALIVE_HTTP is used, the HTTP endpoint MUST be used
// If APP_KEEPALIVE_UDP is used, the UDP endpoint MUST be used
Ping bool `json:"ping,omitempty"`

// App Process ID
// Ping MUST be true if we wish to send a PID
PID uint32 `json:"pid,omitempty"`

// Toggle State
//
// API_TOGGLE_STATE_ENABLE = 1
// API_TOGGLE_STATE_DISABLE = 2
// API_TOGGLE_STATE_RESTART = 3
// API_TOGGLE_STATE_RUNONCE_ENABLE = 4
// API_TOGGLE_STATE_RUNONCE_DISABLE = 5
// API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE = 6
// API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE = 7
//
// API_TOGGLE_STATE_ENABLE: Enable App or Service
// API_TOGGLE_STATE_DISABLE: Disable App or Service
// API_TOGGLE_STATE_RESTART: Restart App or Service, Enable App or Service if Disabled
// API_TOGGLE_STATE_RUNONCE_ENABLE: Enable RunOnce for App or Service
// API_TOGGLE_STATE_RUNONCE_DISABLE: Disable RunOnce for App or Service
// API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE: Enable App or Service and Enable RunOnce
// API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE: Enable App or Service and Disable RunOnce
Toggle uint8 `json:"toggle,omitempty"`

// Return History?
History bool `json:"history,omitempty"`

// KeyValue
KeyValue map[string]interface{} `json:"keyvalue,omitempty"`

// If KeyValueReplace is true, previous KeyValue will be replaced with KeyValue
KeyValueReplace bool `json:"keyvalue-replace,omitempty"`

// Secret is required to access the /api GET and POST endpoints
Secret string `json:"secret,omitempty"`

// CAS IS OPTIONAL
// if CAS is NOT set: we will ignore it and we will override all of our values and state!!!
// if CAS IS SET: we will only override values if our CAS is correct!
// HOWEVER, we will ALWAYS update our PING/LastSeen value REGARDLESS OF CAS!!!
// updating `Ping, LastSeen, or PID` will cause our CAS to be incremented!!!
CAS uint64 `json:"cas,omitempty"`
```


## type API_Response struct {
```golang
// An API Response references our STATE at the time of Request
// If any values change or CAS is incremented, they will STILL reference the premodification state!
//
// When using UDP, we won't be able to respond with all of our data, we're going to have to limit our response size
// We're going to limit our response to: `id, group, pid, started, lastseen, disabled, restart, run-once, shutdown`
// We'll have to ignore `history, keyvalue, and errors`, if they're needed the HTTP endpoint should be used instead

// Unique Identifier
ID string `json:"id,omitempty"`

// Instance ID - UUIDv4 - Only exists IF we're running!
InstanceID string `json:"instance-id,omitempty"`

// Group: `app` or `service`
Group string `json:"group,omitempty"`

// Display Name
Name string `json:"name,omitempty"`

// App Process ID
PID uint32 `json:"pid,omitempty"`

// Timestamp App or Service started at
Started string `json:"started,omitempty"`

// Timestamp App or Service was last seen
LastSeen string `json:"lastseen,omitempty"`

// Is our App or Service Disabled?
Disabled bool `json:"disabled,omitempty"`

// Is our App or Service in a Restart state?
Restart bool `json:"restart,omitempty"`

// Is our App or Service set to RunOnce?
RunOnce bool `json:"run-once,omitempty"`

// Is Patrol in a Shutdown state?
Shutdown bool `json:"shutdown,omitempty"`

// History of previous App or Service states at the time of close()
History []*History `json:"history,omitempty"`

// Current state's KeyValue
KeyValue map[string]interface{} `json:"keyvalue,omitempty"`

// Does this App or Service require a Secret to modify?
Secret bool `json:"secret,omitempty"`

// Did any Errors occur?
Errors []string `json:"errors,omitempty"`

// like all of our other values, CAS is a snapshot of our PREVIOUS state
// we are NEVER going to return our current CAS after modifying our current state or values
// the reason for this is that if a modification request is successful, we know our CAS is CAS + 1
// if we were to take a snapshot, update our object, then get our CAS ---
// we could never actually verify what our current state or values are!!!
// the reason for this has to do with triggers, we NEVER KNOW when we're going to unlock and/or execute triggers!!!
// there are going to be very many scenarios where an API request is made and our CAS is updated more than once!!!
// we're never in a scenario where we take a snapshot, update, and get our CAS WITHOUT UNLOCKING!!!
// if we want to make a clean CAS, we should do a REQUEST without modifying anything(no ping), then do a secondary request without incrementing CAS!
CAS uint64 `json:"cas,omitempty"`

// CASInvalid is the only exception to data that references our previous snapshot
// We need to know if our CAS was successful or not!
// I prefer to have this as invalid and not valid as most requests without a CAS will be valid!
CASInvalid bool `json:"cas-invalid,omitempty"`
```


## type History struct {
```golang
InstanceID string                 `json:"instance-id,omitempty"`
PID        uint32                 `json:"pid,omitempty"`
Started    string                 `json:"started,omitempty"`
LastSeen   string                 `json:"lastseen,omitempty"`
Stopped    string                 `json:"stopped,omitempty"`
Disabled   bool                   `json:"disabled,omitempty"`
Restart    bool                   `json:"restart,omitempty"`
RunOnce    bool                   `json:"run-once,omitempty"`
Shutdown   bool                   `json:"shutdown,omitempty"`
ExitCode   uint8                  `json:"exit-code,omitempty"`
KeyValue   map[string]interface{} `json:"keyvalue,omitempty"`
```
