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
				defer service.mu.Unlock()
				wg.Done()
			}()
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
					log.Printf("./patrol.runServices(): Service ID: %s is not running: \"%s\"\n", id, err)
					// call start trigger
					if service.config.TriggerStart != nil {
						// use goroutine to avoid deadlock
						go service.config.TriggerStart(id, service)
					}
					if err := service.startService(); err != nil {
						log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", id, err)
						// call start failed trigger
						if service.config.TriggerStartFailed != nil {
							// use goroutine to avoid deadlock
							go service.config.TriggerStartFailed(id, service)
						}
					} else {
						log.Printf("./patrol.runServices(): Service ID: %s started\n", id)
						// call started trigger
						if service.config.TriggerStarted != nil {
							// use goroutine to avoid deadlock
							go service.config.TriggerStarted(id, service)
						}
					}
				} else {
					log.Printf("./patrol.runServices(): Service ID: %s is running\n", id)
					// call running trigger
					// this should be thought of a ping/noop
					if service.config.TriggerRunning != nil {
						// use goroutine to avoid deadlock
						go service.config.TriggerRunning(id, service)
					}
				}
			}
		}(id, service)
	}
	wg.Wait()
}
