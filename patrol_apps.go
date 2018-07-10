package patrol

import (
	"log"
	"sync"
	"time"
)

func (self *Patrol) runApps() {
	var wg sync.WaitGroup
	self.mu.RLock()
	// we need a copy of our initial "start" time
	started := self.ticker_running
	shutdown := self.shutdown
	self.mu.RUnlock()
	// we're not going to initially start HTTP and UDP Apps on boot
	// there's a chance these may actually be already running, we're going to wait up to at least PingTimeout * 2
	can_start_pingable := time.Now().After(started.Add(time.Duration(self.config.PingTimeout*2) * time.Second))
	for id, app := range self.apps {
		wg.Add(1)
		go func(id string, app *App) {
			app.mu.Lock()
			defer func() {
				app.mu.Unlock()
				wg.Done()
			}()
			if app.disabled || shutdown {
				// this is disabled
				// check if we're running, if we are we need to shutdown
				if err := app.isAppRunning(); err == nil {
					stop := false
					if self.shutdown {
						// Patrol is shutting down
						log.Printf("./patrol.runApps(): App ID: %s is running AND shutting down! - Signalling!\n", id)
						stop = true
					}
					if !stop && app.disabled {
						// app disabled
						log.Printf("./patrol.runApps(): App ID: %s is running AND disabled! - Signalling!\n", id)
						stop = true
					}
					if stop {
						// signal app to stop
						app.signalStop()
					}
				}
			} else {
				// enabled
				if err := app.isAppRunning(); err != nil {
					// check if we're pingable and if so can we start
					if app.config.KeepAlive == APP_KEEPALIVE_HTTP ||
						app.config.KeepAlive == APP_KEEPALIVE_UDP {
						if !can_start_pingable {
							// we can't start this service yet
							log.Printf("./patrol.runApps(): App ID: %s is Pingable and can't be started yet, ignoring!\n", id)
							return
						}
					}
					// call start trigger
					if app.config.TriggerStart != nil {
						// we're going to temporarily unlock so that we can check our trigger start state externally
						// we're doing this to avoid a deadlock
						// the upside is that the state of our App can't be changed externally in any way to mess up our logic
						app.mu.Unlock()
						app.config.TriggerStart(id, app)
						// and relock since we've deferred
						app.mu.Lock()
						if app.disabled {
							// we can't run this App, we've disabled it externally
							return
						}
						// run!
					}
					log.Printf("./patrol.runApps(): App ID: %s is not running: \"%s\"\n", id, err)
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
