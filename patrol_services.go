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
				// call start trigger
				trigger, _ := self.ServicesTrigger[id]
				if trigger.IsValid() && trigger.Start != nil {
					trigger.Start(id, ps)
				}
				if err := ps.startService(); err != nil {
					log.Printf("./patrol.runServices(): Service ID: %s failed to start: \"%s\"\n", id, err)
					// call start failed trigger
					trigger, _ := self.ServicesTrigger[id]
					if trigger.IsValid() && trigger.StartFailed != nil {
						trigger.StartFailed(id, ps)
					}
				} else {
					log.Printf("./patrol.runServices(): Service ID: %s started\n", id)
					// call started trigger
					trigger, _ := self.ServicesTrigger[id]
					if trigger.IsValid() && trigger.Started != nil {
						trigger.Started(id, ps)
					}
				}
			} else {
				log.Printf("./patrol.runServices(): Service ID: %s is running\n", id)
				// call running trigger
				// this should be thought of a ping/noop
				trigger, _ := self.ServicesTrigger[id]
				if trigger.IsValid() && trigger.Running != nil {
					trigger.Running(id, ps)
				}
			}
		}(id, ps)
	}
	wg.Wait()
}
