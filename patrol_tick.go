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
)

func (self *Patrol) Start() error {
	self.ticker_mu.Lock()
	defer self.ticker_mu.Unlock()
	if self.ticker_running {
		// ticker running
		return ERR_PATROL_ALREADYRUNNING
	}
	go self.tick()
	return nil
}
func (self *Patrol) Stop() error {
	self.ticker_mu.Lock()
	defer self.ticker_mu.Unlock()
	if !self.ticker_running {
		// ticker not running
		return ERR_PATROL_NOTRUNNING
	}
	self.ticker_stop = true
	return nil
}
func (self *Patrol) tick() {
	log.Println("./patrol.tick(): starting")
	self.ticker_mu.Lock()
	if self.ticker_running {
		// ticker is running
		self.ticker_mu.Unlock()
		return
	}
	if self.ticker_stop {
		// ticker is stopped
		self.ticker_mu.Unlock()
		return
	}
	self.ticker_running = true
	self.ticker_mu.Unlock()
	log.Println("./patrol.tick(): started")
	defer func() {
		log.Println("./patrol.tick(): stopping")
		self.ticker_mu.Lock()
		self.ticker_stop = false
		self.ticker_running = false
		self.ticker_mu.Unlock()
		log.Println("./patrol.tick(): stopped")
	}()
	var wg sync.WaitGroup
	for {
		self.ticker_mu.RLock()
		if self.ticker_stop {
			self.ticker_mu.RUnlock()
			// stopped
			return
		}
		self.ticker_mu.RUnlock()
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
		<-time.After(time.Second * time.Duration(self.TickEvery))
	}
}
