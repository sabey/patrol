package patrol

type History struct {
	PID      uint32                 `json:"pid,omitempty"`
	Started  *Timestamp             `json:"started,omitempty"`
	LastSeen *Timestamp             `json:"lastseen,omitempty"`
	Stopped  *Timestamp             `json:"stopped,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
	Restart  bool                   `json:"restart,omitempty"`
	RunOnce  bool                   `json:"run-once,omitempty"`
	Shutdown bool                   `json:"shutdown,omitempty"`
	ExitCode uint8                  `json:"exit-code,omitempty"`
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
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
	h := &History{
		PID:      self.PID,
		Started:  self.Started,
		LastSeen: self.LastSeen,
		Stopped:  self.Stopped,
		Disabled: self.Disabled,
		Restart:  self.Restart,
		RunOnce:  self.RunOnce,
		Shutdown: self.Shutdown,
		ExitCode: self.ExitCode,
		KeyValue: make(map[string]interface{}),
	}
	// dereference
	for k, v := range self.KeyValue {
		h.KeyValue[k] = v
	}
	return h
}
