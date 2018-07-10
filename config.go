package patrol

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	ERR_LISTEN_HTTP_EMPTY       = fmt.Errorf("HTTP Listeners were empty, we required one to exist!")
	ERR_LISTEN_UDP_EMPTY        = fmt.Errorf("UDP Listeners were empty, we required one to exist!")
)

func LoadConfig(
	path string,
) (
	*Config,
	error,
) {
	file, err := os.Open(path)
	if err != nil {
		// couldn't open file
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &Config{}
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
	// we will NOT validate our config here!!!
	// we can validate manually or on create
	return config, nil
}

type Config struct {
	// Apps must contain a unique non empty key: ( 0-9 A-Z a-z - )
	// APP ID MUST be usable as a valid hostname label, ie: len <= 63 AND no starting/ending -
	// Keys are NOT our binary name, Keys are only used as unique identifiers
	// when sending keep alive, this unique identifier WILL be used!
	Apps map[string]*ConfigApp `json:"apps,omitempty"`
	// Services must contain a unique non empty key: ( 0-9 A-Z a-z - )
	// SERVICE ID MUST be usable as a valid hostname label, ie: len <= 63 AND no starting/ending -
	// Keys are NOT our binary name, Keys are only used as unique identifiers
	// when sending keep alive, this unique identifier WILL be used!
	Services map[string]*ConfigService `json:"services,omitempty"`
	// this is an integer value for seconds
	// we will multiply this by time.Second on use
	TickEvery int `json:"tick-every,omitempty"`
	// how many records of history should we store?
	History int `json:"history,omitempty"`
	// used for time.Format()
	// empty defaults to time.String()
	Timestamp string `json:"json-timestamp,omitempty"`
	// we have to have extra configs for our list of listeners - we'll allow multiple ones to exist
	// we won't create any listeners of our own in THIS package!
	// if a HTTP or UDP listener is required, the only time one would be created is in our subpackage `patrol`
	// we would only create one if no other listener was listed
	// once created they will be appended to our listen list
	// https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
	// by default we will listen on :8421 for HTTP and :1248 for UDP
	// neither seem to be used anywhere, good enough for us an easy to remember
	ListenHTTP []string `json:"listen-http,omitempty"`
	ListenUDP  []string `json:"listen-udp,omitempty"`
	// we're going to add options for a default listener here
	// these values won't be used in this package internally
	// IF these values are NIL and we need them, we will listen on LOOPBACK interfaces!!!
	HTTP *ConfigHTTP `json:"http,omitempty"`
	UDP  *ConfigUDP  `json:"udp,omitempty"`
}

func (self *Config) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Config) Clone() *Config {
	if self == nil {
		return nil
	}
	config := &Config{
		Apps:       make(map[string]*ConfigApp),
		Services:   make(map[string]*ConfigService),
		TickEvery:  self.TickEvery,
		History:    self.History,
		Timestamp:  self.Timestamp,
		ListenHTTP: make([]string, 0, len(self.ListenHTTP)),
		ListenUDP:  make([]string, 0, len(self.ListenUDP)),
		HTTP:       self.HTTP.Clone(),
		UDP:        self.UDP.Clone(),
	}
	for k, v := range self.Apps {
		config.Apps[k] = v.Clone()
	}
	for k, v := range self.Services {
		config.Services[k] = v.Clone()
	}
	for _, l := range self.ListenHTTP {
		config.ListenHTTP = append(config.ListenHTTP, l)
	}
	for _, l := range self.ListenUDP {
		config.ListenUDP = append(config.ListenUDP, l)
	}
	return config
}
func (self *Config) Validate() error {
	if len(self.Apps) == 0 &&
		len(self.Services) == 0 {
		// no apps or services found
		return ERR_PATROL_EMPTY
	}
	// we need to check for one exception, JSON keys are case sensitive
	// we won't allow any duplicate case insensitive keys to exist as our ID MAY be used as a hostname label
	// we're actually going to create a secondary dereferenced map with our keys set to lowercase
	// this way we can rely in the future on IDs being lowercase
	apps := make(map[string]*ConfigApp)
	// check apps
	http := false
	udp := false
	for id, app := range self.Apps {
		if id == "" {
			return ERR_APPS_KEY_EMPTY
		}
		if !IsAppServiceID(id) {
			return ERR_APPS_KEY_INVALID
		}
		// dereference
		app = app.Clone()
		if !app.IsValid() {
			return ERR_APPS_APP_NIL
		}
		if err := app.Validate(); err != nil {
			return err
		}
		// create lowercase ID
		id = strings.ToLower(id)
		if _, ok := apps[id]; ok {
			// ID already exists!!
			return ERR_APP_LABEL_DUPLICATE
		}
		apps[id] = app
		if app.KeepAlive == APP_KEEPALIVE_HTTP {
			http = true
		} else if app.KeepAlive == APP_KEEPALIVE_UDP {
			udp = true
		}
	}
	if http && len(self.ListenHTTP) == 0 {
		// no http servers
		return ERR_LISTEN_HTTP_EMPTY
	}
	if udp && len(self.ListenUDP) == 0 {
		// no udp servers
		return ERR_LISTEN_UDP_EMPTY
	}
	// overwrite apps
	self.Apps = apps
	// dereference and lowercase services
	services := make(map[string]*ConfigService)
	// check services
	for id, service := range self.Services {
		if id == "" {
			return ERR_SERVICES_KEY_EMPTY
		}
		if !IsAppServiceID(id) {
			return ERR_SERVICES_KEY_INVALID
		}
		// dereference
		service = service.Clone()
		if !service.IsValid() {
			return ERR_SERVICES_SERVICE_NIL
		}
		if err := service.Validate(); err != nil {
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
	// overwrite services
	self.Services = services
	// config
	if self.TickEvery == 0 {
		self.TickEvery = TICKEVERY_DEFAULT
	} else if self.TickEvery < TICKEVERY_MIN {
		self.TickEvery = TICKEVERY_MIN
	} else if self.TickEvery > TICKEVERY_MAX {
		self.TickEvery = TICKEVERY_MAX
	}
	if self.History == 0 {
		self.History = HISTORY_DEFAULT
	} else if self.History < HISTORY_MIN {
		self.History = HISTORY_MIN
	} else if self.History > HISTORY_MAX {
		self.History = HISTORY_MAX
	}
	return nil
}
