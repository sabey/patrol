package patrol

import (
	"log"
	"sync"
)

func (self *Patrol) runServices() {
	var wg sync.WaitGroup
	for id, ps := range self.Services {
		wg.Add(1)
		go func(id string, ps *PatrolService) {
			defer wg.Done()
			if err := ps.isServiceRunning(); err != nil {
				log.Printf("./patrol.runServices(): Service ID: %s is not running: \"%s\"\n", id, err)
				if err := ps.startService(); err != nil {
					log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", id, err)
				} else {
					log.Printf("./patrol.runServices(): Service ID: %s started\n", id)
				}
			} else {
				log.Printf("./patrol.runServices(): Service ID: %s is running\n", id)
			}
		}(id, ps)
	}
	wg.Wait()
}
