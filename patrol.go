package patrol

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	TICKEVERY_MIN     = 5
	TICKEVERY_MAX     = 120
	TICKEVERY_DEFAULT = 15
)

var (
	ERR_CONFIG_NIL              = fmt.Errorf("Config was NIL")
	ERR_PATROL_EMPTY            = fmt.Errorf("Patrol Apps and Servers were both empty")
	ERR_APPS_KEY_EMPTY          = fmt.Errorf("App Key was empty")
	ERR_APPS_KEY_INVALID        = fmt.Errorf("App Key was invalid")
	ERR_APPS_APP_NIL            = fmt.Errorf("App was nil")
	ERR_SERVICES_KEY_EMPTY      = fmt.Errorf("Service Key was empty")
	ERR_SERVICES_KEY_INVALID    = fmt.Errorf("Service Key was invalid")
	ERR_SERVICES_SERVICE_NIL    = fmt.Errorf("Service was nil")
	ERR_APP_LABEL_DUPLICATE     = fmt.Errorf("Duplicate App Label")
	ERR_SERVICE_LABEL_DUPLICATE = fmt.Errorf("Duplicate Service Label")
)
var (
	// do not change this in this package
	// change it in testing files or for debugging purposes only
	unittesting bool = false
)

func LoadPatrol(
	path string,
) (
	*Patrol,
	error,
) {
	file, err := os.Open(path)
	if err != nil {
		// couldn't open file
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &Patrol{}
	if err := decoder.Decode(config); err != nil {
		// couldn't decode file as json
		return nil, err
	}
	if config == nil {
		// decoded config was nil
		// while this would rarely occur in practice, it is technically possible for config to be nil after decode/unmarshal
		// a json file with the value of "null" would cause this to occur
		// the only time this really plays out in reality is if someone were to POST null to a JSON API
		// this something that is sometimes overlooked
		return nil, ERR_CONFIG_NIL
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type Patrol struct {
	// Apps must contain a unique non empty key: ( 0-9 A-Z a-z - )
	// APP ID MUST be usable as a valid hostname label, ie: len <= 63 AND no starting/ending -
	// Keys are NOT our binary name, Keys are only used as unique identifiers
	// when sending keep alive, this unique identifier WILL be used!
	Apps map[string]*PatrolApp `json:"apps,omitempty"`
	// Services must contain a unique non empty key: ( 0-9 A-Z a-z - )
	// SERVICE ID MUST be usable as a valid hostname label, ie: len <= 63 AND no starting/ending -
	// Keys are NOT our binary name, Keys are only used as unique identifiers
	// when sending keep alive, this unique identifier WILL be used!
	Services map[string]*PatrolService `json:"services,omitempty"`
	// Config
	// this is an integer value for seconds
	// we will multiply this by time.Second on use
	TickEvery int `json:"tick-every,omitempty"`
	// unsafe
	// ticker
	ticker_running bool
	ticker_stop    bool
	ticker_mu      sync.RWMutex
}

func (self *Patrol) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Patrol) IsRunning() bool {
	self.ticker_mu.RLock()
	defer self.ticker_mu.RUnlock()
	return self.ticker_running
}

func (self *Patrol) validate() error {
	if len(self.Apps) == 0 &&
		len(self.Services) == 0 {
		// no apps or services found
		return ERR_PATROL_EMPTY
	}
	// we need to check for one exception, JSON keys are case sensitive
	// we won't allow any duplicate case insensitive keys to exist as our ID MAY be used as a hostname label
	// we're actually going to create a secondary dereferenced map with our keys set to lowercase
	// this way we can rely in the future on IDs being lowercase
	apps := make(map[string]*PatrolApp)
	// check apps
	for id, app := range self.Apps {
		if id == "" {
			return ERR_APPS_KEY_EMPTY
		}
		if !IsAppServiceID(id) {
			return ERR_APPS_KEY_INVALID
		}
		if !app.IsValid() {
			return ERR_APPS_APP_NIL
		}
		if err := app.validate(); err != nil {
			return err
		}
		// create lowercase ID
		id = strings.ToLower(id)
		if _, ok := apps[id]; ok {
			// ID already exists!!
			return ERR_APP_LABEL_DUPLICATE
		}
		apps[id] = app
	}
	// overwrite apps
	self.Apps = apps
	// dereference and lowercase services
	services := make(map[string]*PatrolService)
	// check services
	for id, service := range self.Services {
		if id == "" {
			return ERR_SERVICES_KEY_EMPTY
		}
		if !IsAppServiceID(id) {
			return ERR_SERVICES_KEY_INVALID
		}
		if !service.IsValid() {
			return ERR_SERVICES_SERVICE_NIL
		}
		if err := service.validate(); err != nil {
			return err
		}
		// create lowercase ID
		id = strings.ToLower(id)
		if _, ok := services[id]; ok {
			// ID already exists!!
			return ERR_SERVICE_LABEL_DUPLICATE
		}
		services[id] = service
	}
	self.Services = services
	// config
	if self.TickEvery == 0 {
		self.TickEvery = TICKEVERY_DEFAULT
	} else if self.TickEvery < TICKEVERY_MIN {
		self.TickEvery = TICKEVERY_MIN
	} else if self.TickEvery > TICKEVERY_MAX {
		self.TickEvery = TICKEVERY_MAX
	}
	return nil
}
