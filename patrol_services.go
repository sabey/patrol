package main

import (
	"log"
	"sync"
)

func (self *Patrol) runServices() {
	var wg sync.WaitGroup
	for service, ps := range self.Services {
		wg.Add(1)
		go func(service string, ps *PatrolService) {
			defer wg.Done()
			if err := ps.isServiceRunning(service); err != nil {
				log.Printf("the service %s is not running: \"%s\"\n", service, err)
				if err := ps.startService(service); err != nil {
					log.Printf("the service %s failed to start: \"%s\"\n", service, err)
				} else {
					log.Printf("the service %s started\n", service)
				}
			} else {
				log.Printf("the service %s is running\n", service)
			}
		}(service, ps)
	}
	wg.Wait()
}
