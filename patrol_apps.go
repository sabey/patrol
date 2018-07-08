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
			app.mu.Lock()
			defer func() {
				defer app.mu.Unlock()
				wg.Done()
			}()
			if app.disabled || self.shutdown {
				// this is disabled
				// check if we're running, if we are we need to shutdown
				if err := app.isAppRunning(); err == nil {
					if self.shutdown {
						// Patrol is shutting down
						log.Printf("./patrol.runApps(): App ID: %s is running AND shutting down! - Signalling!\n", id)
					} else {
						// app disabled
						log.Printf("./patrol.runApps(): App ID: %s is running AND disabled! - Signalling!\n", id)
					}
					// signal app to stop
					app.signalStop()
				}
			} else {
				// enabled
				if err := app.isAppRunning(); err != nil {
					log.Printf("./patrol.runApps(): App ID: %s is not running: \"%s\"\n", id, err)
					// call start trigger
					if app.config.TriggerStart != nil {
						// use goroutine to avoid deadlock
						go app.config.TriggerStart(id, app)
					}
					if err := app.startApp(); err != nil {
						log.Printf("./patrol.runApps(): App ID: %s failed to start: \"%s\"\n", id, err)
						// call start failed trigger
						if app.config.TriggerStartFailed != nil {
							// use goroutine to avoid deadlock
							go app.config.TriggerStartFailed(id, app)
						}
					} else {
						log.Printf("./patrol.runApps(): App ID: %s started\n", id)
						// call started trigger
						if app.config.TriggerStarted != nil {
							// use goroutine to avoid deadlock
							go app.config.TriggerStarted(id, app)
						}
					}
				} else {
					log.Printf("./patrol.runApps(): App ID: %s is running\n", id)
					// call running trigger
					// this should be thought of a ping/noop
					if app.config.TriggerRunning != nil {
						// use goroutine to avoid deadlock
						go app.config.TriggerRunning(id, app)
					}
				}
			}
		}(id, app)
	}
	wg.Wait()
}
