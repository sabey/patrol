package patrol

type ConfigService struct {
	// management method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your app may function
	Management int `json:"management,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
	Name string `json:"name,omitempty"`
	// Service is the name of the executable and/or parameter
	Service string `json:"service,omitempty"`
	// these are a list of valid exit codes to ignore when returned from "service * status"
	// by default 0 is always ignored, it is assumed to mean that the service is running
	IgnoreExitCodes []uint8 `json:"ignore-exit-codes,omitempty"`
	// these are NOT supported with JSON for obvious reasons
	// these will have to be set manually!!!
	// Triggers
	TriggerStart func(
		id string,
		app *Service,
	) `json:"-"`
	TriggerStarted func(
		id string,
		app *Service,
	) `json:"-"`
	TriggerStartFailed func(
		id string,
		app *Service,
	) `json:"-"`
	TriggerRunning func(
		id string,
		app *Service,
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
		TriggerStart:       self.TriggerStart,
		TriggerStarted:     self.TriggerStarted,
		TriggerStartFailed: self.TriggerStartFailed,
		TriggerRunning:     self.TriggerRunning,
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
