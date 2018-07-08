package patrol

import (
	"log"
	"sync"
)

func (self *Patrol) runApps() {
	var wg sync.WaitGroup
	for id, app := range self.apps {
		wg.Add(1)
		go func(id string, app *App) {
			defer wg.Done()
			if err := app.isAppRunning(); err != nil {
				log.Printf("./patrol.runApps(): App ID: %s is not running: \"%s\"\n", id, err)
				// call start trigger
				if app.config.TriggerStart != nil {
					app.config.TriggerStart(id, app)
				}
				if err := app.startApp(); err != nil {
					log.Printf("./patrol.runApps(): App ID: %s failed to start: \"%s\"\n", id, err)
					// call start failed trigger
					if app.config.TriggerStartFailed != nil {
						app.config.TriggerStartFailed(id, app)
					}
				} else {
					log.Printf("./patrol.runApps(): App ID: %s started\n", id)
					// call started trigger
					if app.config.TriggerStarted != nil {
						app.config.TriggerStarted(id, app)
					}
				}
			} else {
				log.Printf("./patrol.runApps(): App ID: %s is running\n", id)
				// call running trigger
				// this should be thought of a ping/noop
				if app.config.TriggerRunning != nil {
					app.config.TriggerRunning(id, app)
				}
			}
		}(id, app)
	}
	wg.Wait()
}
