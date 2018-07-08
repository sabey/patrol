package patrol

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
				// call start trigger
				trigger, _ := self.AppsTrigger[id]
				if trigger.IsValid() && trigger.Start != nil {
					trigger.Start(id, pa)
				}
				if err := pa.startApp(); err != nil {
					log.Printf("./patrol.runApps(): App ID: %s failed to start: \"%s\"\n", id, err)
					// call start failed trigger
					trigger, _ := self.AppsTrigger[id]
					if trigger.IsValid() && trigger.StartFailed != nil {
						trigger.StartFailed(id, pa)
					}
				} else {
					log.Printf("./patrol.runApps(): App ID: %s started\n", id)
					// call started trigger
					trigger, _ := self.AppsTrigger[id]
					if trigger.IsValid() && trigger.Started != nil {
						trigger.Started(id, pa)
					}
				}
			} else {
				log.Printf("./patrol.runApps(): App ID: %s is running\n", id)
				// call running trigger
				// this should be thought of a ping/noop
				trigger, _ := self.AppsTrigger[id]
				if trigger.IsValid() && trigger.Running != nil {
					trigger.Running(id, pa)
				}
			}
		}(id, pa)
	}
	wg.Wait()
}
