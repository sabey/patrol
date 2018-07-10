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
	keyvalue map[string]interface{}
	history  []*History
	started  time.Time
	lastseen time.Time
	disabled bool // takes its initial value from config
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
	self.disabled = true
}
func (self *Service) Enable() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.disabled = false
}
func (self *Service) GetStarted() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.started
}
func (self *Service) GetLastSeen() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.lastseen
}
func (self *Service) GetKeyValue() map[string]interface{} {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.getKeyValue()
}
func (self *Service) getKeyValue() map[string]interface{} {
	// dereference
	kv := make(map[string]interface{})
	for k, v := range self.keyvalue {
		kv[k] = v
	}
	return kv
}
func (self *Service) SetKeyValue(
	kv map[string]interface{},
) {
	self.mu.Lock()
	for k, v := range kv {
		self.keyvalue[k] = v
	}
	self.mu.Unlock()
}
func (self *Service) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.mu.Lock()
	self.keyvalue = make(map[string]interface{})
	// dereference
	for k, v := range kv {
		self.keyvalue[k] = v
	}
	self.mu.Unlock()
}
func (self *Service) GetHistory() []*History {
	self.mu.RLock()
	defer self.mu.RUnlock()
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
			Stopped: &Timestamp{
				Time: time.Now(),
				f:    self.patrol.config.Timestamp,
			},
			Shutdown: self.patrol.shutdown,
			KeyValue: self.getKeyValue(),
		}
		if !self.started.IsZero() {
			h.Started = &Timestamp{
				Time: self.started,
				f:    self.patrol.config.Timestamp,
			}
		}
		if !self.lastseen.IsZero() {
			h.LastSeen = &Timestamp{
				Time: self.lastseen,
				f:    self.patrol.config.Timestamp,
			}
		}
		self.history = append(self.history, h)
		// reset values
		self.started = time.Time{}
		if self.config.KeyValueClear {
			// clear keyvalues
			self.keyvalue = make(map[string]interface{})
		}
		// call trigger in a go routine so we don't deadlock
		if self.config.TriggerClosed != nil {
			go self.config.TriggerClosed(self.id, self, h)
		}
	}
}
func (self *Service) startService() error {
	// close service
	self.close()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	m := self.config.GetManagementStart()
	p := self.config.GetManagementStartParameter()
	if m == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, p)
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), p)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	var cmd *exec.Cmd
	m := self.config.GetManagementStatus()
	p := self.config.GetManagementStatusParameter()
	if m == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, p)
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), p)
	}
	if err := cmd.Run(); err != nil {
		// NOT running!
		// close service
		self.close()
		return err
	}
	// running!
	now := time.Now()
	if self.started.IsZero() {
		// we need to set started since this is our first time seeing this service
		self.started = now
	}
	self.lastseen = now
	return nil
}
func (self *Service) stopService() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	m := self.config.GetManagementStop()
	p := self.config.GetManagementStopParameter()
	if m == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, p)
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), p)
	}
	if err := cmd.Run(); err != nil {
		// failed to stop
		return err
	}
	// stopped!
	// close service
	self.close()
	return nil
}
