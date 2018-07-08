package patrol

import (
	"fmt"
)

var (
	ERR_SERVICE_EMPTY              = fmt.Errorf("Service was empty")
	ERR_SERVICE_MAXLENGTH          = fmt.Errorf("Service was longer than 63 bytes")
	ERR_SERVICE_NAME_EMPTY         = fmt.Errorf("Service Name was empty")
	ERR_SERVICE_NAME_MAXLENGTH     = fmt.Errorf("Service Name was longer than 255 bytes")
	ERR_SERVICE_MANAGEMENT_INVALID = fmt.Errorf("Service Management was invalid, please select a method!")
	ERR_SERVICE_INVALID_EXITCODE   = fmt.Errorf("Service contained an Invalid Exit Code")
	ERR_SERVICE_DUPLICATE_EXITCODE = fmt.Errorf("Service contained a Duplicate Exit Code")
)

type ConfigService struct {
	// management method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your service may function
	Management int `json:"management,omitempty"`
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
	// these are NOT supported with JSON for obvious reasons
	// these will have to be set manually!!!
	// Triggers
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
		Management:         self.Management,
		Name:               self.Name,
		Service:            self.Service,
		IgnoreExitCodes:    make([]uint8, 0, len(self.IgnoreExitCodes)),
		Disabled:           self.Disabled,
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
	if self.Management < SERVICE_MANAGEMENT_SERVICE ||
		self.Management > SERVICE_MANAGEMENT_INITD {
		// unknown management value
		return ERR_SERVICE_MANAGEMENT_INVALID
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
