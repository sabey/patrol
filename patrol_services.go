package patrol

import (
	"log"
	"sync"
)

func (self *Patrol) runServices() {
	var wg sync.WaitGroup
	// we're going to ignore any shutdown checks in this function
	// we're only interested in the state of shutdown for Services as we're responsible for running AND managing state
	for _, service := range self.services {
		wg.Add(1)
		go func(service *Service) {
			defer wg.Done()
			// we're not going to defer unlocking our service mutex, we're going to occasionally unlock and allow our tiggers to run
			// for example when we check if our service is running, if we call close() we want to trigger our close right away
			// if we do not unlock, we could then call startService() without having ever signalled our close trigger
			service.mu.Lock()
			// we have to check if we're running on every loop, regardless if we're disabled
			// there's a chance our service could become enabled should we check and find we're running
			// if we aren't running and we call close() and call our close trigger
			is_running_err := service.isServiceRunning()
			is_running := (is_running_err == nil)
			// we need to run our state triggers
			if is_running {
				// we're running!
				log.Printf("./patrol.runServices(): Service ID: %s is running\n", service.id)
				if service.config.TriggerRunning != nil {
					service.mu.Unlock()
					service.config.TriggerRunning(service)
					service.mu.Lock()
				}
				// if we're disabled or restarting we're going to signal our services to stop
				if service.restart {
					// signal our service to restart
					log.Printf("./patrol.runServices(): Service ID: %s is running AND we're restarting! - Restarting!\n", service.id)
					if err := service.restartService(); err != nil {
						log.Printf("./patrol.runServices(): Service ID: %s failed to restart: \"%s\"\n", service.id, err)
					} else {
						log.Printf("./patrol.runServices(): Service ID: %s restarted\n", service.id)
					}
					// we will only attempt to restart ONCE, we consume restart even if we fail to restart!
					service.restart = false
					// it's going to ultimately be up to our Service to exit
					// we're not going to immediately attempt to start our app on this tick
					// in fact, if our app chooses not to exit we will do nothing!
				} else if service.disabled {
					// signal our service to stop
					log.Printf("./patrol.runServices(): Service ID: %s is running AND is disabled! - Stopping!\n", service.id)
					if err := service.stopService(); err != nil {
						log.Printf("./patrol.runServices(): Service ID: %s failed to stop: \"%s\"\n", service.id, err)
					} else {
						log.Printf("./patrol.runServices(): Service ID: %s stopped\n", service.id)
					}
				}
				service.mu.Unlock()
				// we're done!
				return
			} else {
				// we aren't running
				if service.disabled {
					// service is disabled
					log.Printf("./patrol.runServices(): Service ID: %s is not running AND is disabled! - Reason: \"%s\"\n", service.id, is_running_err)
					if service.config.TriggerDisabled != nil {
						service.mu.Unlock()
						service.config.TriggerDisabled(service)
						service.mu.Lock()
					}
					// check if we're still disabled
					if service.disabled {
						// still disabled!!!
						// nothing we can do
						service.mu.Unlock()
						// we're done!
						return
					}
					// we're now enabled!!!
					log.Printf("./patrol.runServices(): Service ID: %s was disabled and now enabled!\n", service.id)
				} else {
					// service is enabled and we aren't running
					log.Printf("./patrol.runServices(): Service ID: %s was not running, starting! - Reason: \"%s\"\n", service.id, is_running_err)
				}
			}
			// time to start our service!
			log.Printf("./patrol.runServices(): Service ID: %s starting!\n", service.id)
			if service.config.TriggerStart != nil {
				service.mu.Unlock()
				service.config.TriggerStart(service)
				service.mu.Lock()
				// this will be our LAST chance to check disabled!!
				if service.disabled {
					// disabled!!!
					service.mu.Unlock()
					// we're done!
					log.Printf("./patrol.runServices(): Service ID: %s can't start, we're disabled!\n", service.id)
					return
				}
			}
			// run!
			if err := service.startService(); err != nil {
				log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", service.id, err)
				// call start failed trigger
				if service.config.TriggerStartFailed != nil {
					service.mu.Unlock()
					// we're done!
					service.config.TriggerStartFailed(service)
					return
				}
			} else {
				log.Printf("./patrol.runServices(): Service ID: %s started\n", service.id)
				// call started trigger
				if service.config.TriggerStarted != nil {
					service.mu.Unlock()
					// we're done!
					service.config.TriggerStarted(service)
					return
				}
			}
			service.mu.Unlock()
			// we're done!
			return
		}(service)
	}
	wg.Wait()
}
