package main

import (
	"sync"
)

func (self *Patrol) runApps() {
	var wg sync.WaitGroup
	for app, pa := range self.Apps {
		wg.Add(1)
		go func(app string, pa *PatrolApp) {
			defer wg.Done()
			//is the app running?
			//~~try to run it
			//~~~~it failed
			//~~~~it succeeded
			//~~its running so done

		}(app, pa)

	}
	wg.Wait()

}
