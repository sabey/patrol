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

var (
	ERR_SERVICE_EMPTY              = fmt.Errorf("Service was empty")
	ERR_SERVICE_MAXLENGTH          = fmt.Errorf("Service was longer than 63 bytes")
	ERR_SERVICE_NAME_EMPTY         = fmt.Errorf("Service Name was empty")
	ERR_SERVICE_NAME_MAXLENGTH     = fmt.Errorf("Service Name was longer than 255 bytes")
	ERR_SERVICE_MANAGEMENT_INVALID = fmt.Errorf("Service Management was invalid, please select a method!")
	ERR_SERVICE_INVALID_EXITCODE   = fmt.Errorf("Service contained an Invalid Exit Code")
	ERR_SERVICE_DUPLICATE_EXITCODE = fmt.Errorf("Service contained a Duplicate Exit Code")
)

type Service struct {
	// safe
	patrol *Patrol
	id     string // we want a reference to our parent ID
	config *ConfigService
	// unsafe
	history []*History
	started time.Time
	mu      sync.RWMutex
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
func (self *Service) GetHistory() []*History {
	// dereference
	history := make([]*History, 0, len(self.history))
	for _, h := range self.history {
		history = append(history, h.clone())
	}
	return history
}
func (self *Service) saveHistory(
	shutdown bool,
) {
	if !self.started.IsZero() {
		// save history
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			Started:  self.started,
			Stopped:  time.Now(),
			Shutdown: shutdown,
		}
		self.history = append(self.history, h)
		// unset previous started so we don't create duplicate histories
		self.started = time.Time{}
		// call trigger in a go routine so we don't deadlock
		if self.config.TriggerStopped != nil {
			go self.config.TriggerStopped(self.id, self, h)
		}
	}
}
func (self *Service) startService() error {
	// we are ASSUMING our service isn't started!!!
	// this function should only ever be called by tick()
	// we gotta lock and defer to set history
	self.mu.Lock()
	defer self.mu.Unlock()
	// save history
	self.saveHistory(false)
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
		// save history
		self.saveHistory(false)
		return err
	}
	// running!
	return nil
}
