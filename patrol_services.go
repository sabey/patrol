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
				// this is disabled
				// check if we're running, if we are we need to shutdown
				if err := service.isServiceRunning(); err == nil {
					log.Printf("./patrol.runServices(): Service ID: %s is running AND disabled!\n", id)
					// shut it down
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
