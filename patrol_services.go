package patrol

import (
	"log"
	"sync"
)

func (self *Patrol) runServices() {
	var wg sync.WaitGroup
	for id, service := range self.services {
		if service.IsDisabled() {
			// ignore
			continue
		}
		wg.Add(1)
		go func(id string, service *Service) {
			defer wg.Done()
			if err := service.isServiceRunning(); err != nil {
				log.Printf("./patrol.runServices(): Service ID: %s is not running: \"%s\"\n", id, err)
				// call start trigger
				if service.config.TriggerStart != nil {
					service.config.TriggerStart(id, service)
				}
				if err := service.startService(); err != nil {
					log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", id, err)
					// call start failed trigger
					if service.config.TriggerStartFailed != nil {
						service.config.TriggerStartFailed(id, service)
					}
				} else {
					log.Printf("./patrol.runServices(): Service ID: %s started\n", id)
					// call started trigger
					if service.config.TriggerStarted != nil {
						service.config.TriggerStarted(id, service)
					}
				}
			} else {
				log.Printf("./patrol.runServices(): Service ID: %s is running\n", id)
				// call running trigger
				// this should be thought of a ping/noop
				if service.config.TriggerRunning != nil {
					service.config.TriggerRunning(id, service)
				}
			}
		}(id, service)
	}
	wg.Wait()
}
