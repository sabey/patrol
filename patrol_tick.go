package patrol

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ERR_PATROL_ALREADYRUNNING = fmt.Errorf("Patrol is already running!")
	ERR_PATROL_NOTRUNNING     = fmt.Errorf("Patrol is NOT running!")
	ERR_PATROL_SHUTDOWN       = fmt.Errorf("Patrol is Shutdown")
)

func (self *Patrol) Start() error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.shutdown {
		// patrol is shutdown
		return ERR_PATROL_SHUTDOWN
	}
	if !self.ticker_running.IsZero() {
		// ticker running
		return ERR_PATROL_ALREADYRUNNING
	}
	go self.tick()
	return nil
}
func (self *Patrol) Stop() error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.ticker_running.IsZero() {
		// ticker not running
		return ERR_PATROL_NOTRUNNING
	}
	self.ticker_stop = true
	return nil
}
func (self *Patrol) tick() {
	log.Println("./patrol.tick(): starting")
	self.mu.Lock()
	if !self.ticker_running.IsZero() {
		// ticker is running
		self.mu.Unlock()
		return
	}
	if self.shutdown {
		// patrol is shutdown
		self.mu.Unlock()
		return
	}
	if self.ticker_stop {
		// ticker is stopped
		self.mu.Unlock()
		return
	}
	self.ticker_running = time.Now()
	self.mu.Unlock()
	// signal that we've started
	if self.config.TriggerStarted != nil {
		// since we're not in a lock we're going to wait for this function
		self.config.TriggerStarted(self)
	}
	log.Println("./patrol.tick(): started")
	defer func() {
		log.Println("./patrol.tick(): stopping")
		// signal our trigger we've stopped
		// we should signal our trigger before we signal all of our apps
		if self.config.TriggerStopped != nil {
			// since we're not in a lock we're going to wait for this function
			self.config.TriggerStopped(self)
		}
		// we need to signal to all of our apps that we're stopping!
		self.signalStopApps()
		// close our ticker
		self.mu.Lock()
		self.ticker_stop = false
		self.ticker_running = time.Time{}
		self.mu.Unlock()
		log.Println("./patrol.tick(): stopped")
	}()
	var wg sync.WaitGroup
	for {
		// call tick
		if self.config.TriggerTick != nil {
			self.config.TriggerTick(self)
		}
		self.mu.RLock()
		// if we're shutting down, do not close yet
		// we first want to notify our apps/services we're closing
		shutdown := self.shutdown
		if self.ticker_stop {
			self.mu.RUnlock()
			// stopped
			return
		}
		self.mu.RUnlock()
		// tick
		wg.Add(2)
		go func() {
			defer wg.Done()
			self.runApps()
		}()
		go func() {
			defer wg.Done()
			self.runServices()
		}()
		wg.Wait()
		if shutdown {
			// we're done!
			return
		}
		<-time.After(time.Second * time.Duration(self.config.TickEvery))
	}
}
