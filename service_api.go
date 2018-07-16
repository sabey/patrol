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
	if request.Toggle > 0 {
		self.toggle(request.Toggle)
	}
}
func (self *Service) toggle(
	toggle uint8,
) {
	if toggle == API_TOGGLE_STATE_ENABLE {
		self.disabled = false
	} else if toggle == API_TOGGLE_STATE_DISABLE {
		self.disabled = true
		self.restart = false
		self.run_once = false
		self.run_once_consumed = false
	} else if toggle == API_TOGGLE_STATE_RESTART {
		self.disabled = false
		self.restart = true
		self.run_once = false
		self.run_once_consumed = false
	} else if toggle == API_TOGGLE_STATE_RUNONCE_ENABLE {
		self.run_once = true
		if !self.started.IsZero() {
			// we're already running, we must consume run_once
			self.run_once_consumed = true
		}
	} else if toggle == API_TOGGLE_STATE_RUNONCE_DISABLE {
		self.run_once = false
		self.run_once_consumed = false
	} else if toggle == API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE {
		self.disabled = false
		self.run_once = true
		if !self.started.IsZero() {
			// we're already running, we must consume run_once
			self.run_once_consumed = true
		}
	} else if toggle == API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE {
		self.disabled = false
		self.run_once = false
		self.run_once_consumed = false
	}
}
func (self *Service) Snapshot() *API_Response {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.apiResponse(api_endpoint_snapshot)
}
func (self *Service) apiResponse(
	endpoint uint8,
) *API_Response {
	result := &API_Response{
		Disabled: self.disabled,
		Restart:  self.restart,
		RunOnce:  self.run_once,
	}
	if endpoint != api_endpoint_status {
		// we don't need these values for individual status objects
		result.ID = self.id
		result.Group = "service"
		// we need to read lock patrol
		self.patrol.mu.RLock()
		result.Shutdown = self.patrol.shutdown
		self.patrol.mu.RUnlock()
	}
	if endpoint != api_endpoint_udp {
		// we don't want either of these for UDP, it's too much data
		//
		// we're not going to include history in our snapshot, it's too much data to dereference
		// if history is needed it should be taken additionally after snapshot
		if endpoint != api_endpoint_snapshot {
			result.History = self.getHistory()
		}
		result.KeyValue = self.getKeyValue()
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
