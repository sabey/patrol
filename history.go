package patrol

import (
	"time"
)

type History struct {
	PID      uint32    `json:"pid,omitempty"`
	Started  time.Time `json:"started,omitempty"`
	Stopped  time.Time `json:"stopped,omitempty"`
	Disabled bool      `json:"disabled,omitempty"`
	Shutdown bool      `json:"shutdown,omitempty"`
	ExitCode uint8     `json:"exit-code,omitempty"`
}

func (self *History) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *History) clone() *History {
	if self == nil {
		return nil
	}
	return &History{
		PID:      self.PID,
		Started:  self.Started,
		Stopped:  self.Stopped,
		Disabled: self.Disabled,
		Shutdown: self.Shutdown,
		ExitCode: self.ExitCode,
	}
}
