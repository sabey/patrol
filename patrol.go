package patrol

import (
	"sync"
	"time"
)

const (
	TICKEVERY_MIN     = 5
	TICKEVERY_MAX     = 120
	TICKEVERY_DEFAULT = 15
	HISTORY_MIN       = 5
	HISTORY_MAX       = 1000
	HISTORY_DEFAULT   = 100
)

var (
	// do not change this in this package
	// change it in testing files or for debugging purposes only
	unittesting bool = false
)

func CreatePatrol(
	config *Config,
) (
	*Patrol,
	error,
) {
	// deference
	config = config.Clone()
	if !config.IsValid() {
		return nil, ERR_CONFIG_NIL
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	p := &Patrol{
		config:   config,
		apps:     make(map[string]*App),
		services: make(map[string]*Service),
	}
	for id, app := range config.Apps {
		p.apps[id] = &App{
			id:     id,
			patrol: p,
			config: app,
		}
	}
	for id, service := range config.Services {
		p.services[id] = &Service{
			id:     id,
			patrol: p,
			config: service,
		}
	}
	return p, nil
}

type Patrol struct {
	// safe
	config   *Config
	apps     map[string]*App
	services map[string]*Service
	// unsafe
	// ticker
	ticker_running time.Time
	ticker_stop    bool
	ticker_mu      sync.RWMutex
}

func (self *Patrol) IsValid() bool {
	return self.config.IsValid()
}
func (self *Patrol) IsRunning() bool {
	self.ticker_mu.RLock()
	defer self.ticker_mu.RUnlock()
	return !self.ticker_running.IsZero()
}
func (self *Patrol) GetStarted() time.Time {
	self.ticker_mu.RLock()
	defer self.ticker_mu.RUnlock()
	return self.ticker_running
}
func (self *Patrol) GetConfig() *Config {
	return self.config.Clone()
}
func (self *Patrol) GetApps() map[string]*App {
	// derefence
	apps := make(map[string]*App)
	for k, v := range self.apps {
		apps[k] = v
	}
	return apps
}
func (self *Patrol) GetServices() map[string]*Service {
	// derefence
	services := make(map[string]*Service)
	for k, v := range self.services {
		services[k] = v
	}
	return services
}
