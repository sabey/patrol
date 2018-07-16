package patrol

import (
	"log"
	"sync"
	"time"
)

func (self *Patrol) signalStopApps() {
	var wg sync.WaitGroup
	log.Printf("./patrol.signalStopApps(): signalling to all apps that we are stopping!\n")
	for _, app := range self.apps {
		wg.Add(1)
		go func(app *App) {
			defer wg.Done()
			app.mu.Lock()
			if app.isAppRunning() == nil {
				log.Printf("./patrol.signalStopApps(): App ID: %s is running - Signalling!\n", app.id)
				app.signalStop()
			}
			app.mu.Unlock()
		}(app)
	}
	wg.Wait()
}
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
	for _, app := range self.apps {
		wg.Add(1)
		go func(app *App) {
			defer wg.Done()
			// we're not going to defer unlocking our app mutex, we're going to occasionally unlock and allow our tiggers to run
			// for example when we check if our app is running, if we call close() we want to trigger our close right away
			// if we do not unlock, we could then call startApp() without having ever signalled our close trigger
			app.mu.Lock()
			// we have to check if we're running on every loop, regardless if we're disabled
			// there's a chance our app could become enabled should we check and find we're running
			// if we aren't running and we call close() and call our close trigger
			is_running_err := app.isAppRunning()
			is_running := (is_running_err == nil)
			// if we're shutting down we're going to ignore state triggers and signal apps to stop
			if shutdown {
				// we're shutting down!
				if is_running {
					// signal our app to stop
					log.Printf("./patrol.runApps(): App ID: %s is running AND we're shutting down! - Signalling!\n", app.id)
					app.signalStop()
				}
				app.mu.Unlock()
				// we're done!
				return
			}
			// patrol is still active!
			// we need to run our state triggers
			if is_running {
				// we're running!
				log.Printf("./patrol.runApps(): App ID: %s is running\n", app.id)
				if app.config.TriggerRunning != nil {
					app.mu.Unlock()
					app.config.TriggerRunning(app)
					app.mu.Lock()
				}
				// if we're disabled or restarting we're going to signal our apps to stop
				if app.restart {
					// signal our app to stop
					log.Printf("./patrol.runApps(): App ID: %s is running AND we're restarting! - Signalling!\n", app.id)
					app.signalStop()
					// we will only attempt to restart ONCE, we consume restart even if we fail to restart!
					app.restart = false
					// it's going to ultimately be up to our App to exit
					// we're not going to immediately attempt to start our app on this tick
					// in fact, if our app chooses not to exit we will do nothing!
				} else if app.disabled {
					// signal our app to stop
					log.Printf("./patrol.runApps(): App ID: %s is running AND is disabled! - Signalling!\n", app.id)
					app.signalStop()
				}
				app.mu.Unlock()
				// we're done!
				return
			} else {
				// we aren't running
				if app.disabled {
					// app is disabled
					log.Printf("./patrol.runApps(): App ID: %s is not running AND is disabled! - Reason: \"%s\"\n", app.id, is_running_err)
					if app.config.TriggerDisabled != nil {
						app.mu.Unlock()
						app.config.TriggerDisabled(app)
						app.mu.Lock()
					}
					// check if we're still disabled
					if app.disabled {
						// still disabled!!!
						// nothing we can do
						app.mu.Unlock()
						// we're done!
						return
					}
					// we're now enabled!!!
					log.Printf("./patrol.runApps(): App ID: %s was disabled and now enabled!\n", app.id)
				} else {
					// app is enabled and we aren't running
					log.Printf("./patrol.runApps(): App ID: %s was not running, starting! - Reason: \"%s\"\n", app.id, is_running_err)
				}
			}
			// check if we're pingable and if we can start yet
			if !is_running {
				if app.config.KeepAlive == APP_KEEPALIVE_HTTP ||
					app.config.KeepAlive == APP_KEEPALIVE_UDP {
					if !can_start_pingable {
						// we can't start this service yet
						app.mu.Unlock()
						// we're done!
						log.Printf("./patrol.runApps(): App ID: %s is Pingable and can't be started yet, ignoring!\n", app.id)
						return
					}
				}
			}
			// time to start our app!
			log.Printf("./patrol.runApps(): App ID: %s starting!\n", app.id)
			if app.config.TriggerStart != nil {
				app.mu.Unlock()
				app.config.TriggerStart(app)
				app.mu.Lock()
				// this will be our LAST chance to check disabled!!
				if app.disabled {
					// disabled!!!
					app.mu.Unlock()
					// we're done!
					log.Printf("./patrol.runApps(): App ID: %s can't start, we're disabled!\n", app.id)
					return
				}
			}
			// run!
			if err := app.startApp(); err != nil {
				log.Printf("./patrol.runApps(): App ID: %s failed to start: \"%s\"\n", app.id, err)
				// call start failed trigger
				if app.config.TriggerStartFailed != nil {
					app.mu.Unlock()
					// we're done!
					app.config.TriggerStartFailed(app)
					return
				}
			} else {
				log.Printf("./patrol.runApps(): App ID: %s started\n", app.id)
				// call started trigger
				if app.config.TriggerStarted != nil {
					app.mu.Unlock()
					// we're done!
					app.config.TriggerStarted(app)
					return
				}
			}
			app.mu.Unlock()
			// we're done!
			return
		}(app)
	}
	wg.Wait()
}
