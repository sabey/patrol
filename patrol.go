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
			id:       id,
			patrol:   p,
			config:   app,
			disabled: app.Disabled,
			keyvalue: make(map[string]interface{}),
		}
	}
	for id, service := range config.Services {
		p.services[id] = &Service{
			id:       id,
			patrol:   p,
			config:   service,
			disabled: service.Disabled,
			keyvalue: make(map[string]interface{}),
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
	shutdown bool
	// ticker
	ticker_running time.Time
	ticker_stop    bool
	mu             sync.RWMutex
}

func (self *Patrol) IsValid() bool {
	return self.config.IsValid()
}
func (self *Patrol) IsRunning() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return !self.ticker_running.IsZero()
}
func (self *Patrol) GetStarted() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
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
func (self *Patrol) Shutdown() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.shutdown = true
}
func (self *Patrol) IsShutdown() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.shutdown
}
