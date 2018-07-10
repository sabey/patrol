package patrol

type ConfigHTTP struct {
	Listen string `json:"listen,omitempty"`
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
	}
	return config
}

type ConfigUDP struct {
	Listen string `json:"listen,omitempty"`
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
	}
	return config
}
