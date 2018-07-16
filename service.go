package patrol

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
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
	// disabled takes its initial value from our config
	disabled bool
	// restart will ALWAYS cause our service to become enabled!
	restart bool
	// run once has to be consumed
	// once our service is running we will instantly consume runonce, once stopped we will become disabled
	// if our service is stopped, we will start and consume runonce
	run_once          bool
	run_once_consumed bool
	mu                sync.RWMutex
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
func (self *Service) IsDisabled() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.disabled
}
func (self *Service) IsRestart() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.restart
}
func (self *Service) IsRunOnce() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.run_once
}
func (self *Service) IsRunOnceConsumed() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.run_once_consumed
}
func (self *Service) Toggle(
	toggle uint8,
) {
	self.mu.Lock()
	self.toggle(toggle)
	self.mu.Unlock()
}
func (self *Service) Enable() {
	self.mu.Lock()
	self.toggle(API_TOGGLE_STATE_ENABLE)
	self.mu.Unlock()
}
func (self *Service) Disable() {
	self.mu.Lock()
	self.toggle(API_TOGGLE_STATE_DISABLE)
	self.mu.Unlock()
}
func (self *Service) Restart() {
	self.mu.Lock()
	self.toggle(API_TOGGLE_STATE_RESTART)
	self.mu.Unlock()
}
func (self *Service) EnableRunOnce() {
	self.mu.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_ENABLE)
	self.mu.Unlock()
}
func (self *Service) DisableRunOnce() {
	self.mu.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_DISABLE)
	self.mu.Unlock()
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
	return self.getHistory()
}
func (self *Service) getHistory() []*History {
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
			Disabled: self.disabled,
			Restart:  self.restart,
			// we want to know if we CONSUMED run_once, not if run_once is currently true!!!
			RunOnce:  self.run_once_consumed,
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
		self.lastseen = time.Time{}
		if self.run_once_consumed {
			// we have to disable our app!
			self.toggle(API_TOGGLE_STATE_DISABLE)
		}
		if self.config.KeyValueClear {
			// clear keyvalues
			self.keyvalue = make(map[string]interface{})
		}
		// we're not going to use a goroutine here
		// we're assumed to be in a lock
		// we're going to unlock and then relock so that we can call our trigger
		if self.config.TriggerClosed != nil {
			self.mu.Unlock()
			self.config.TriggerClosed(self, h)
			self.mu.Lock()
		}
	}
}
func (self *Service) startService() error {
	now := time.Now()
	// consume restart
	self.restart = false
	// consume runonce
	if self.run_once {
		self.run_once_consumed = true
	}
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
	// check exit code
	if err := cmd.Run(); err != nil {
		f := false
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows.
			// Although package syscall is generally platform dependent,
			// WaitStatus is defined for both Unix and Windows and in both cases has an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := uint8(status.ExitStatus())
				for _, i := range self.config.IgnoreExitCodesStart {
					if i == exit_code {
						// error code is ignored
						f = true
						break
					}
				}
			}
		}
		if !f {
			// unknown error
			// DO NOT CLOSE, WE'RE UNSURE IF WE'VE RESTARTED!!!
			return err
		}
	}
	// exit code 0
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
	// check exit code
	if err := cmd.Run(); err != nil {
		f := false
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows.
			// Although package syscall is generally platform dependent,
			// WaitStatus is defined for both Unix and Windows and in both cases has an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := uint8(status.ExitStatus())
				for _, i := range self.config.IgnoreExitCodesStatus {
					if i == exit_code {
						// error code is ignored
						f = true
						break
					}
				}
			}
		}
		if !f {
			// unknown error
			// close service
			self.close()
			return err
		}
	}
	// running!
	now := time.Now()
	if self.started.IsZero() {
		// Service was not running
		self.started = now
		// we need to call our started trigger
		if self.config.TriggerStarted != nil {
			self.mu.Unlock()
			self.config.TriggerStarted(self)
			self.mu.Lock()
		}
	} else {
		// service was previously started
		self.lastseen = now
	}
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
	// check exit code
	if err := cmd.Run(); err != nil {
		f := false
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows.
			// Although package syscall is generally platform dependent,
			// WaitStatus is defined for both Unix and Windows and in both cases has an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := uint8(status.ExitStatus())
				for _, i := range self.config.IgnoreExitCodesStop {
					if i == exit_code {
						// error code is ignored
						f = true
						break
					}
				}
			}
		}
		if !f {
			// unknown error
			// DO NOT CLOSE, WE'RE UNSURE IF WE'RE STOPPED!!!
			return err
		}
	}
	// stopped!
	// close service
	self.close()
	return nil
}
func (self *Service) restartService() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	m := self.config.GetManagementRestart()
	p := self.config.GetManagementRestartParameter()
	if m == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, p)
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), p)
	}
	// check exit code
	if err := cmd.Run(); err != nil {
		f := false
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			// This works on both Unix and Windows.
			// Although package syscall is generally platform dependent,
			// WaitStatus is defined for both Unix and Windows and in both cases has an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit_code := uint8(status.ExitStatus())
				for _, i := range self.config.IgnoreExitCodesRestart {
					if i == exit_code {
						// error code is ignored
						f = true
						break
					}
				}
			}
		}
		if !f {
			// unknown error
			// DO NOT CLOSE, WE'RE UNSURE IF WE'VE RESTARTED!!!
			return err
		}
	}
	// restarted!
	// close service
	self.close()
	return nil
}
