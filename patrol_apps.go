package main

import (
	"log"
	"sync"
)

func (self *Patrol) runApps() {
	var wg sync.WaitGroup
	for id, pa := range self.Apps {
		wg.Add(1)
		go func(id string, pa *PatrolApp) {
			defer wg.Done()
			if err := pa.isAppRunning(); err != nil {
				log.Printf("./patrol.runApps(): App ID: %s is not running: \"%s\"\n", id, err)
				if err := pa.startApp(); err != nil {
					log.Printf("./patrol.runApps(): App ID: %s failed to start: \"%s\"\n", id, err)
				} else {
					log.Printf("./patrol.runApps(): App ID: %s started\n", id)
				}
			} else {
				log.Printf("./patrol.runApps(): App ID: %s is running\n", id)
			}
		}(id, pa)
	}
	wg.Wait()
}
