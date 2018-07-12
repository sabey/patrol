package patrol

func (self *Service) apiRequest(
	request *API_Request,
) {
	// keyvalue
	if request.KeyValueReplace {
		// replace
		// dereference
		kv := make(map[string]interface{})
		for k, v := range request.KeyValue {
			kv[k] = v
		}
		self.keyvalue = kv
	} else {
		// merge
		for k, v := range request.KeyValue {
			self.keyvalue[k] = v
		}
	}
	// toggle
	if request.Toggle == API_TOGGLE_ENABLE {
		self.disabled = false
	} else if request.Toggle == API_TOGGLE_DISABLE {
		self.disabled = true
	} else if request.Toggle == API_TOGGLE_RUNONCE {
		// TODO
	}
}
func (self *Service) apiResponse(
	endpoint uint8,
) *API_Response {
	result := &API_Response{
		Disabled: self.disabled,
	}
	if endpoint != api_endpoint_udp {
		// we don't need these values for individual Services
		result.ID = self.id
		result.Group = "service"
		result.History = self.getHistory()
		result.KeyValue = self.getKeyValue()
	}
	if endpoint == api_endpoint_status {
		// we need to read lock patrol
		self.patrol.mu.RLock()
		result.Shutdown = self.patrol.shutdown
		self.patrol.mu.RUnlock()
	}
	if !self.started.IsZero() {
		result.Started = &Timestamp{
			Time: self.started,
			f:    self.patrol.config.Timestamp,
		}
	}
	if !self.lastseen.IsZero() {
		result.LastSeen = &Timestamp{
			Time: self.lastseen,
			f:    self.patrol.config.Timestamp,
		}
	}
	return result
}
