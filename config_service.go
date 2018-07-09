package patrol

import (
	"fmt"
)

var (
	ERR_SERVICE_EMPTY                     = fmt.Errorf("Service was empty")
	ERR_SERVICE_MAXLENGTH                 = fmt.Errorf("Service was longer than 63 bytes")
	ERR_SERVICE_NAME_EMPTY                = fmt.Errorf("Service Name was empty")
	ERR_SERVICE_NAME_MAXLENGTH            = fmt.Errorf("Service Name was longer than 255 bytes")
	ERR_SERVICE_MANAGEMENT_INVALID        = fmt.Errorf("Service Management was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_START_INVALID  = fmt.Errorf("Service Management Start was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_STATUS_INVALID = fmt.Errorf("Service Management Status was invalid, please select a method!")
	ERR_SERVICE_MANAGEMENT_STOP_INVALID   = fmt.Errorf("Service Management Stop was invalid, please select a method!")
	ERR_SERVICE_INVALID_EXITCODE          = fmt.Errorf("Service contained an Invalid Exit Code")
	ERR_SERVICE_DUPLICATE_EXITCODE        = fmt.Errorf("Service contained a Duplicate Exit Code")
)

type ConfigService struct {
	// management method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your service may function
	// we're not going to have a single management method, we're going to use different ones for start/status/stop
	// IF management is set, it will take priority over the other methods
	Management       int `json:"management,omitempty"`
	ManagementStart  int `json:"management-start,omitempty"`
	ManagementStatus int `json:"management-status,omitempty"`
	ManagementStop   int `json:"management-stop,omitempty"`
	// we're going to allow different start/status/stop parameters to be overwritten, if they are empty they will use the default value
	ManagementStartParameter  string `json:"management-start-parameter,omitempty"`
	ManagementStatusParameter string `json:"management-status-parameter,omitempty"`
	ManagementStopParameter   string `json:"management-stop-parameter,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
	Name string `json:"name,omitempty"`
	// Service is the name of the executable and/or parameter
	Service string `json:"service,omitempty"`
	// these are a list of valid exit codes to ignore when returned from "service * status"
	// by default 0 is always ignored, it is assumed to mean that the service is running
	IgnoreExitCodes []uint8 `json:"ignore-exit-codes,omitempty"`
	// if Disabled is true the Service won't be executed until enabled
	// the only way to enable this once loaded is to use an API or restart Patrol
	// if Disabled is true the Service MAY be running, we will just avoid watching it!
	Disabled bool `json:"disabled,omitempty"`
	// should we stop this service once patrol is shutting down?
	// if patrol crashes we won't be able to garauntee this is stopped
	StopOnShutdown bool `json:"stop-on-shutdown,omitempty"`
	// clear keyvalue on new instance?
	KeyValueClear bool `json:"keyvalue-clear,omitempty"`
	// these are NOT supported with JSON for obvious reasons
	// these will have to be set manually!!!
	// Triggers
	// we're going to allow our Start function to overwrite if our Service is Disabled
	// this will be just incase we want to hold a disabled state outside of this service, such as in a database, just incase we crash
	// we'll check the value of Service.disabled on return
	TriggerStart func(
		id string,
		service *Service,
	) `json:"-"`
	TriggerStarted func(
		id string,
		service *Service,
	) `json:"-"`
	TriggerStartFailed func(
		id string,
		service *Service,
	) `json:"-"`
	TriggerRunning func(
		id string,
		service *Service,
	) `json:"-"`
	TriggerClosed func(
		id string,
		service *Service,
		history *History,
	) `json:"-"`
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
		Management:                self.Management,
		ManagementStart:           self.ManagementStart,
		ManagementStatus:          self.ManagementStatus,
		ManagementStop:            self.ManagementStop,
		ManagementStartParameter:  self.ManagementStartParameter,
		ManagementStatusParameter: self.ManagementStatusParameter,
		ManagementStopParameter:   self.ManagementStopParameter,
		Name:               self.Name,
		Service:            self.Service,
		IgnoreExitCodes:    make([]uint8, 0, len(self.IgnoreExitCodes)),
		Disabled:           self.Disabled,
		StopOnShutdown:     self.StopOnShutdown,
		KeyValueClear:      self.KeyValueClear,
		TriggerStart:       self.TriggerStart,
		TriggerStarted:     self.TriggerStarted,
		TriggerStartFailed: self.TriggerStartFailed,
		TriggerRunning:     self.TriggerRunning,
		TriggerClosed:      self.TriggerClosed,
	}
	for _, i := range self.IgnoreExitCodes {
		config.IgnoreExitCodes = append(config.IgnoreExitCodes, i)
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
	exists := make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodes {
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
