package patrol

import (
	"time"
)

type History struct {
	Started  time.Time `json:"started,omitempty"`
	Stopped  time.Time `json:"stopped,omitempty"`
	Shutdown bool      `json:"shutdown,omitempty"` // were we responsible for terminating this process?
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
		Started:  self.Started,
		Stopped:  self.Stopped,
		Shutdown: self.Shutdown,
	}
}
