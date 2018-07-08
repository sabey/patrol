package patrol

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const (
	// service name maximum length in bytes
	SERVICE_NAME_MAXLENGTH = 255
	// service maximum length in bytes
	SERVICE_MAXLENGTH = 63
)

const (
	// this is an alias for "service * status"
	// ie: "service ssh status"
	SERVICE_MANAGEMENT_SERVICE = iota + 1
	// this is an alias for "/etc/init.d/* status"
	// ie: "/etc/init.d/ssh status"
	SERVICE_MANAGEMENT_INITD
)

type Service struct {
	// safe
	patrol *Patrol
	id     string // we want a reference to our parent ID
	config *ConfigService
	// unsafe
	history  []*History
	started  time.Time
	disabled bool // takes its initial value from config
	shutdown bool
	mu       sync.RWMutex
}

func (self *Service) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Service) GetID() string {
	return self.id
}
func (self *Service) GetPatrol() *Patrol {
	return self.patrol
}
func (self *Service) GetConfig() *ConfigService {
	return self.config.Clone()
}
func (self *Service) IsRunning() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return !self.started.IsZero()
}
func (self *Service) IsDisabled() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.disabled
}
func (self *Service) Disable() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if !self.started.IsZero() {
		// shutdown this process
		self.shutdown = true
	}
	self.disabled = true
}
func (self *Service) Enable() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.shutdown = false
	self.disabled = false
}
func (self *Service) GetStarted() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.started
}
func (self *Service) GetHistory() []*History {
	// dereference
	history := make([]*History, 0, len(self.history))
	for _, h := range self.history {
		history = append(history, h.clone())
	}
	return history
}
func (self *Service) close() {
	if !self.started.IsZero() {
		// close service
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			Started:  self.started,
			Stopped:  time.Now(),
			Shutdown: self.shutdown,
		}
		self.history = append(self.history, h)
		// unset previous started so we don't create duplicate histories
		self.started = time.Time{}
		// call trigger in a go routine so we don't deadlock
		if self.config.TriggerClosed != nil {
			go self.config.TriggerClosed(self.id, self, h)
		}
	}
}
func (self *Service) startService() error {
	// we are ASSUMING our service isn't started!!!
	// this function should only ever be called by tick()
	// we gotta lock and defer to set history
	self.mu.Lock()
	defer self.mu.Unlock()
	// close service
	self.close()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	if self.config.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, "start")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), "start")
	}
	if err := cmd.Run(); err != nil {
		// failed to start
		return err
	}
	// started!
	self.started = now
	return nil
}
func (self *Service) isServiceRunning() error {
	// if we determine a process is NOT running we will set history - we will NOT attempt to restart anything!
	// lock and defer incase we have to set history!
	self.mu.Lock()
	defer self.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	var cmd *exec.Cmd
	if self.config.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, "status")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), "status")
	}
	if err := cmd.Run(); err != nil {
		// NOT running!
		// close service
		self.close()
		return err
	}
	// running!
	if self.started.IsZero() {
		// we need to set started since this is our first time seeing this service
		self.started = time.Now()
	}
	return nil
}
