package patrol

import (
	"context"
	"fmt"
	"os/exec"
	"sabey.co/patrol/cas"
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
	// instance ID only exists IF we're running!
	instance_id string
	// history will wrap our cas Objects Lock/RLock mutex
	// history is NOT included in our cas Object because we didn't want to restructure Patrol
	history []*History
	o       *cas.Service
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
func (self *Service) GetInstanceID() string {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.instance_id
}
func (self *Service) GetPatrol() *Patrol {
	return self.patrol
}
func (self *Service) GetConfig() *ConfigService {
	return self.config.Clone()
}
func (self *Service) GetCAS() uint64 {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetCAS()
}
func (self *Service) IsRunning() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return !self.o.GetStarted().IsZero()
}
func (self *Service) GetStarted() time.Time {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetStarted()
}
func (self *Service) GetLastSeen() time.Time {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetLastSeen()
}
func (self *Service) IsDisabled() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsDisabled()
}
func (self *Service) IsRestart() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRestart()
}
func (self *Service) IsRunOnce() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRunOnce()
}
func (self *Service) IsRunOnceConsumed() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRunOnceConsumed()
}
func (self *Service) Toggle(
	toggle uint8,
) {
	self.o.Lock()
	self.toggle(toggle)
	self.o.Unlock()
}
func (self *Service) Enable() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_ENABLE)
	self.o.Unlock()
}
func (self *Service) Disable() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_DISABLE)
	self.o.Unlock()
}
func (self *Service) Restart() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RESTART)
	self.o.Unlock()
}
func (self *Service) EnableRunOnce() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_ENABLE)
	self.o.Unlock()
}
func (self *Service) DisableRunOnce() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_DISABLE)
	self.o.Unlock()
}
func (self *Service) GetKeyValue() map[string]interface{} {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetKeyValue()
}
func (self *Service) SetKeyValue(
	kv map[string]interface{},
) {
	self.o.Lock()
	self.o.SetKeyValue(kv)
	self.o.Unlock()
}
func (self *Service) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.o.Lock()
	self.o.ReplaceKeyValue(kv)
	self.o.Unlock()
}
func (self *Service) GetHistory() []*History {
	self.o.RLock()
	defer self.o.RUnlock()
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
	if !self.o.GetStarted().IsZero() {
		// save history
		self.o.Increment() // we have to increment for modifying History
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			InstanceID: self.instance_id,
			Stopped: &Timestamp{
				Time:            time.Now(),
				TimestampFormat: self.patrol.config.Timestamp,
			},
			Disabled: self.o.IsDisabled(),
			Restart:  self.o.IsRestart(),
			// we want to know if we CONSUMED run_once, not if run_once is currently true!!!
			RunOnce:  self.o.IsRunOnceConsumed(),
			Shutdown: self.patrol.shutdown,
			KeyValue: self.o.GetKeyValue(),
		}
		if !self.o.GetStarted().IsZero() {
			h.Started = &Timestamp{
				Time:            self.o.GetStarted(),
				TimestampFormat: self.patrol.config.Timestamp,
			}
		}
		if !self.o.GetLastSeen().IsZero() {
			h.LastSeen = &Timestamp{
				Time:            self.o.GetLastSeen(),
				TimestampFormat: self.patrol.config.Timestamp,
			}
		}
		self.history = append(self.history, h)
		// reset values
		self.instance_id = ""
		self.o.SetStarted(time.Time{})
		self.o.SetLastSeen(time.Time{})
		if self.o.IsRunOnceConsumed() {
			// we have to disable our app!
			self.toggle(API_TOGGLE_STATE_DISABLE)
		}
		if self.config.KeyValueClear {
			// clear keyvalues
			self.o.ReplaceKeyValue(nil)
		}
		// we're not going to use a goroutine here
		// we're assumed to be in a lock
		// we're going to unlock and then relock so that we can call our trigger
		if self.config.TriggerClosed != nil {
			self.o.Unlock()
			self.config.TriggerClosed(self, h)
			self.o.Lock()
		}
	}
}
func (self *Service) startService() error {
	now := time.Now()
	// consume restart
	self.o.SetRestart(false)
	// consume runonce
	if self.o.IsRunOnce() {
		self.o.SetRunOnceConsumed(true)
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
	self.instance_id = uuidMust(uuidV4())
	self.o.SetStarted(now)
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
	if self.o.GetStarted().IsZero() {
		// Service was not running
		self.instance_id = uuidMust(uuidV4())
		self.o.SetStarted(now)
		// we need to call our started trigger
		if self.config.TriggerStarted != nil {
			self.o.Unlock()
			self.config.TriggerStarted(self)
			self.o.Lock()
		}
	} else {
		// service was previously started
		self.o.SetLastSeen(now)
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
