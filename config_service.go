package patrol

import (
	"encoding/json"
	"fmt"
)

var (
	ERR_SERVICE_EMPTY                      = fmt.Errorf("Service was empty")
	ERR_SERVICE_MAXLENGTH                  = fmt.Errorf("Service was longer than 63 bytes")
	ERR_SERVICE_NAME_EMPTY                 = fmt.Errorf("Service Name was empty")
	ERR_SERVICE_NAME_MAXLENGTH             = fmt.Errorf("Service Name was longer than 255 bytes")
	ERR_SERVICE_MANAGEMENT_INVALID         = fmt.Errorf("Service Management was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_START_INVALID   = fmt.Errorf("Service Management Start was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_STATUS_INVALID  = fmt.Errorf("Service Management Status was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_STOP_INVALID    = fmt.Errorf("Service Management Stop was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_RESTART_INVALID = fmt.Errorf("Service Management Restart was invalid, please select a method!")
	ERR_SERVICE_INVALID_EXITCODE           = fmt.Errorf("Service contained an Invalid Exit Code")
	ERR_SERVICE_DUPLICATE_EXITCODE         = fmt.Errorf("Service contained a Duplicate Exit Code")
)

type ConfigService struct {
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
	//
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
	// Extra Unstructured Data
	X json.RawMessage `json:"x,omitempty"`
}

func (self *ConfigService) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *ConfigService) Clone() *ConfigService {
	if self == nil {
		return nil
	}
	config := &ConfigService{
		Management:                 self.Management,
		ManagementStart:            self.ManagementStart,
		ManagementStatus:           self.ManagementStatus,
		ManagementStop:             self.ManagementStop,
		ManagementRestart:          self.ManagementRestart,
		ManagementStartParameter:   self.ManagementStartParameter,
		ManagementStatusParameter:  self.ManagementStatusParameter,
		ManagementStopParameter:    self.ManagementStopParameter,
		ManagementRestartParameter: self.ManagementRestartParameter,
		Name:                   self.Name,
		Service:                self.Service,
		IgnoreExitCodesStart:   make([]uint8, 0, len(self.IgnoreExitCodesStart)),
		IgnoreExitCodesStatus:  make([]uint8, 0, len(self.IgnoreExitCodesStatus)),
		IgnoreExitCodesStop:    make([]uint8, 0, len(self.IgnoreExitCodesStop)),
		IgnoreExitCodesRestart: make([]uint8, 0, len(self.IgnoreExitCodesRestart)),
		Disabled:               self.Disabled,
		KeyValue:               make(map[string]interface{}),
		KeyValueClear:          self.KeyValueClear,
		Secret:                 self.Secret,
		TriggerStart:           self.TriggerStart,
		TriggerStarted:         self.TriggerStarted,
		TriggerStartFailed:     self.TriggerStartFailed,
		TriggerRunning:         self.TriggerRunning,
		TriggerDisabled:        self.TriggerDisabled,
		TriggerClosed:          self.TriggerClosed,
		X:                      dereference(self.X),
	}
	for k, v := range self.KeyValue {
		config.KeyValue[k] = v
	}
	for _, i := range self.IgnoreExitCodesStart {
		config.IgnoreExitCodesStart = append(config.IgnoreExitCodesStart, i)
	}
	for _, i := range self.IgnoreExitCodesStatus {
		config.IgnoreExitCodesStatus = append(config.IgnoreExitCodesStatus, i)
	}
	for _, i := range self.IgnoreExitCodesStop {
		config.IgnoreExitCodesStop = append(config.IgnoreExitCodesStop, i)
	}
	for _, i := range self.IgnoreExitCodesRestart {
		config.IgnoreExitCodesRestart = append(config.IgnoreExitCodesRestart, i)
	}
	return config
}
func (self *ConfigService) Validate() error {
	if self.Management == 0 &&
		(self.ManagementStart > 0 ||
			self.ManagementStatus > 0 ||
			self.ManagementStop > 0) {
		// use specific management values
		// start
		if self.ManagementStart < SERVICE_MANAGEMENT_SERVICE ||
			self.ManagementStart > SERVICE_MANAGEMENT_INITD {
			// unknown management value
			return ERR_SERVICE_MANAGEMENT_START_INVALID
		}
		// status
		if self.ManagementStatus < SERVICE_MANAGEMENT_SERVICE ||
			self.ManagementStatus > SERVICE_MANAGEMENT_INITD {
			// unknown management value
			return ERR_SERVICE_MANAGEMENT_STATUS_INVALID
		}
		// stop
		if self.ManagementStop < SERVICE_MANAGEMENT_SERVICE ||
			self.ManagementStop > SERVICE_MANAGEMENT_INITD {
			// unknown management value
			return ERR_SERVICE_MANAGEMENT_STOP_INVALID
		}
		// restart
		if self.ManagementRestart < SERVICE_MANAGEMENT_SERVICE ||
			self.ManagementRestart > SERVICE_MANAGEMENT_INITD {
			// unknown management value
			return ERR_SERVICE_MANAGEMENT_RESTART_INVALID
		}
	} else {
		// use master value
		if self.Management < SERVICE_MANAGEMENT_SERVICE ||
			self.Management > SERVICE_MANAGEMENT_INITD {
			// unknown management value
			return ERR_SERVICE_MANAGEMENT_INVALID
		}
	}
	if self.Service == "" {
		return ERR_SERVICE_EMPTY
	}
	if len(self.Service) > SERVICE_MAXLENGTH {
		return ERR_SERVICE_MAXLENGTH
	}
	if self.Name == "" {
		return ERR_SERVICE_NAME_EMPTY
	}
	if len(self.Name) > SERVICE_NAME_MAXLENGTH {
		return ERR_SERVICE_NAME_MAXLENGTH
	}
	if len(self.Secret) > SECRET_MAX_LENGTH {
		return ERR_SECRET_TOOLONG
	}
	// start
	exists := make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodesStart {
		if ec == 0 {
			// can't ignore 0
			return ERR_SERVICE_INVALID_EXITCODE
		}
		if _, ok := exists[ec]; ok {
			// exit code already exists
			return ERR_SERVICE_DUPLICATE_EXITCODE
		}
		// does not exist
		exists[ec] = struct{}{}
	}
	// status
	exists = make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodesStatus {
		if ec == 0 {
			// can't ignore 0
			return ERR_SERVICE_INVALID_EXITCODE
		}
		if _, ok := exists[ec]; ok {
			// exit code already exists
			return ERR_SERVICE_DUPLICATE_EXITCODE
		}
		// does not exist
		exists[ec] = struct{}{}
	}
	// stop
	exists = make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodesStop {
		if ec == 0 {
			// can't ignore 0
			return ERR_SERVICE_INVALID_EXITCODE
		}
		if _, ok := exists[ec]; ok {
			// exit code already exists
			return ERR_SERVICE_DUPLICATE_EXITCODE
		}
		// does not exist
		exists[ec] = struct{}{}
	}
	// restart
	exists = make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodesRestart {
		if ec == 0 {
			// can't ignore 0
			return ERR_SERVICE_INVALID_EXITCODE
		}
		if _, ok := exists[ec]; ok {
			// exit code already exists
			return ERR_SERVICE_DUPLICATE_EXITCODE
		}
		// does not exist
		exists[ec] = struct{}{}
	}
	return nil
}
func (self *ConfigService) GetManagementStart() int {
	if self.Management > 0 {
		return self.Management
	}
	return self.ManagementStart
}
func (self *ConfigService) GetManagementStatus() int {
	if self.Management > 0 {
		return self.Management
	}
	return self.ManagementStatus
}
func (self *ConfigService) GetManagementStop() int {
	if self.Management > 0 {
		return self.Management
	}
	return self.ManagementStop
}
func (self *ConfigService) GetManagementRestart() int {
	if self.Management > 0 {
		return self.Management
	}
	return self.ManagementRestart
}
func (self *ConfigService) GetManagementStartParameter() string {
	if self.ManagementStartParameter == "" {
		return "start"
	}
	return self.ManagementStartParameter
}
func (self *ConfigService) GetManagementStatusParameter() string {
	if self.ManagementStatusParameter == "" {
		return "status"
	}
	return self.ManagementStatusParameter
}
func (self *ConfigService) GetManagementStopParameter() string {
	if self.ManagementStopParameter == "" {
		return "stop"
	}
	return self.ManagementStopParameter
}
func (self *ConfigService) GetManagementRestartParameter() string {
	if self.ManagementRestartParameter == "" {
		return "restart"
	}
	return self.ManagementRestartParameter
}
