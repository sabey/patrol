package patrol

import (
	"encoding/json"
)

type ConfigHTTP struct {
	Listen string `json:"listen,omitempty"`
	// Extra Unstructured Data
	X json.RawMessage `json:"x,omitempty"`
	// TODO: https?
}

func (self *ConfigHTTP) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *ConfigHTTP) Clone() *ConfigHTTP {
	if self == nil {
		return nil
	}
	config := &ConfigHTTP{
		Listen: self.Listen,
		X:      dereference(self.X),
	}
	return config
}

type ConfigUDP struct {
	Listen string `json:"listen,omitempty"`
	// Extra Unstructured Data
	X json.RawMessage `json:"x,omitempty"`
}

func (self *ConfigUDP) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *ConfigUDP) Clone() *ConfigUDP {
	if self == nil {
		return nil
	}
	config := &ConfigUDP{
		Listen: self.Listen,
		X:      dereference(self.X),
	}
	return config
}
