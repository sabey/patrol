package patrol

import (
	"log"
	"sync"
)

func (self *Patrol) runServices() {
	var wg sync.WaitGroup
	self.mu.RLock()
	shutdown := self.shutdown
	self.mu.RUnlock()
	for id, service := range self.services {
		wg.Add(1)
		go func(id string, service *Service) {
			service.mu.Lock()
			defer func() {
				service.mu.Unlock()
				wg.Done()
			}()
			if service.config.TriggerDisabled != nil &&
				service.disabled && !shutdown {
				// we're going to temporarily unlock so that we can check our trigger disabled state externally
				// we're doing this to avoid a deadlock
				// the upside is that the state of our App can't be changed externally in any way to mess up our logic
				service.mu.Unlock()
				service.config.TriggerDisabled(service)
				// and relock since we've deferred
				service.mu.Lock()
				// we can continue to check our state now
			}
			if service.disabled || shutdown {
				// we're either disabled or shutting down
				// check if we're running, if we are we need to shutdown
				if err := service.isServiceRunning(); err == nil {
					// shut it down
					stop := false
					if self.shutdown {
						// Patrol is shutting down
						if service.config.StopOnShutdown {
							// service should be stopped on shutdown
							log.Printf("./patrol.runServices(): Service ID: %s is running AND shutting down! - Stopping!\n", id)
							stop = true
						} // ignore
					}
					// check if we're disabled
					if !stop && service.disabled {
						// service disabled
						log.Printf("./patrol.runServices(): Service ID: %s is running AND disabled! - Stopping!\n", id)
						stop = true
					}
					if stop {
						// there's no triggers for this
						// once a service is closed we will use that as a trigger
						if err := service.stopService(); err != nil {
							log.Printf("./patrol.runServices(): Service ID: %s failed to stop: \"%s\"\n", id, err)
						} else {
							log.Printf("./patrol.runServices(): Service ID: %s stopped\n", id)
						}
					}
				}
			} else {
				// enabled
				if err := service.isServiceRunning(); err != nil {
					// call start trigger
					if service.config.TriggerStart != nil {
						// we're going to temporarily unlock so that we can check our trigger start state externally
						// we're doing this to avoid a deadlock
						// the upside is that the state of our Service can't be changed externally in any way to mess up our logic
						service.mu.Unlock()
						service.config.TriggerStart(service)
						// and relock since we've deferred
						service.mu.Lock()
						if service.disabled {
							// we can't run this Service, we've disabled it externally
							return
						}
						// run!
					}
					log.Printf("./patrol.runServices(): Service ID: %s is not running: \"%s\"\n", id, err)
					if err := service.startService(); err != nil {
						log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", id, err)
						// call start failed trigger
						if service.config.TriggerStartFailed != nil {
							// use goroutine to avoid deadlock
							go service.config.TriggerStartFailed(service)
						}
					} else {
						log.Printf("./patrol.runServices(): Service ID: %s started\n", id)
						// call started trigger
						if service.config.TriggerStarted != nil {
							// use goroutine to avoid deadlock
							go service.config.TriggerStarted(service)
						}
					}
				} else {
					log.Printf("./patrol.runServices(): Service ID: %s is running\n", id)
					// call running trigger
					// this should be thought of a ping/noop
					if service.config.TriggerRunning != nil {
						// use goroutine to avoid deadlock
						go service.config.TriggerRunning(service)
					}
				}
			}
		}(id, service)
	}
	wg.Wait()
}
