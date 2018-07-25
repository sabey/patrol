package patrol

import (
	"log"
	"os"
	"sabey.co/patrol/cas"
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

const (
	PATROL_ENV_UNITTEST_KEY   = `PATROL_UNITTEST`
	PATROL_ENV_UNITTEST_VALUE = `I KNOW WHAT I AM DOING`
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
	// we're going to check if we're unittesting
	// this isn't the ideal way to do this, but it will work
	if os.Getenv(PATROL_ENV_UNITTEST_KEY) == PATROL_ENV_UNITTEST_VALUE {
		log.Println("./patrol.CreatePatrol(): We are running in unittesting mode!")
		// unittesting!!!
		p.config.unittesting = true
		// we're going to override tickevery so we don't have to wait ages for unittests
		p.config.TickEvery = 2
		// we won't start HTTP/UDP apps until PingTimeout * 2 on boot
		// we're going to override that it's at least something low
		p.config.PingTimeout = 3
	}
	// add apps
	for id, app := range config.Apps {
		p.apps[id] = &App{
			id:     id,
			patrol: p,
			config: app,
			o:      cas.CreateApp(app.Disabled),
		}
		// add preexisting keyvalues
		p.apps[id].ReplaceKeyValue(app.KeyValue)
	}
	// add services
	for id, service := range config.Services {
		p.services[id] = &Service{
			id:     id,
			patrol: p,
			config: service,
			o:      cas.CreateService(service.Disabled),
		}
		// add preexisting keyvalues
		p.services[id].ReplaceKeyValue(service.KeyValue)
	}
	if config.TriggerStart != nil {
		// start patrol
		// we do NOT need to use a goroutine
		// we'll use this as additional validation
		// if this returns an error we WONT return a patrol object
		if err := config.TriggerStart(p); err != nil {
			// failed to trigger start
			return nil, err
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
	if self == nil {
		return false
	}
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
func (self *Patrol) GetApp(
	key string,
) *App {
	return self.apps[key]
}
func (self *Patrol) GetServices() map[string]*Service {
	// derefence
	services := make(map[string]*Service)
	for k, v := range self.services {
		services[k] = v
	}
	return services
}
func (self *Patrol) GetService(
	key string,
) *Service {
	return self.services[key]
}
func (self *Patrol) Shutdown() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if !self.shutdown && self.config.TriggerShutdown != nil {
		// never shutdown
		// use goroutine to avoid deadlock
		go self.config.TriggerShutdown(self)
	}
	self.shutdown = true
	self.stop()
}
func (self *Patrol) IsShutdown() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.shutdown
}
