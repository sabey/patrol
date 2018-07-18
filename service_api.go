package patrol

func (self *Service) apiRequest(
	request *API_Request,
) bool {
	// CAS Attributes:
	//
	cas_valid := true
	if request.CAS > 0 &&
		request.CAS != self.o.GetCAS() {
		// CAS does not match!
		cas_valid = false
	}
	if cas_valid {
		// keyvalue
		if request.KeyValueReplace {
			// replace
			self.o.ReplaceKeyValue(request.KeyValue)
		} else {
			// merge
			if len(request.KeyValue) > 0 {
				self.o.SetKeyValue(request.KeyValue)
			}
		}
		// toggle
		if request.Toggle > 0 {
			self.toggle(request.Toggle)
		}
	}
	// Non-CAS Attributes:
	return cas_valid
}
func (self *Service) toggle(
	toggle uint8,
) {
	if toggle == API_TOGGLE_STATE_ENABLE {
		self.o.SetDisabled(false)
	} else if toggle == API_TOGGLE_STATE_DISABLE {
		self.o.SetDisabled(true)
		self.o.SetRestart(false)
		self.o.SetRunOnce(false)
		self.o.SetRunOnceConsumed(false)
	} else if toggle == API_TOGGLE_STATE_RESTART {
		self.o.SetDisabled(false)
		self.o.SetRestart(true)
		self.o.SetRunOnce(false)
		self.o.SetRunOnceConsumed(false)
	} else if toggle == API_TOGGLE_STATE_RUNONCE_ENABLE {
		self.o.SetRunOnce(true)
		if !self.o.GetStarted().IsZero() {
			// we're already running, we must consume run_once
			self.o.SetRunOnceConsumed(true)
		}
	} else if toggle == API_TOGGLE_STATE_RUNONCE_DISABLE {
		self.o.SetRunOnce(false)
		self.o.SetRunOnceConsumed(false)
	} else if toggle == API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE {
		self.o.SetDisabled(false)
		self.o.SetRunOnce(true)
		if !self.o.GetStarted().IsZero() {
			// we're already running, we must consume run_once
			self.o.SetRunOnceConsumed(true)
		}
	} else if toggle == API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE {
		self.o.SetDisabled(false)
		self.o.SetRunOnce(false)
		self.o.SetRunOnceConsumed(false)
	}
}
func (self *Service) Snapshot() *API_Response {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.apiResponse(api_endpoint_snapshot)
}
func (self *Service) apiResponse(
	endpoint uint8,
) *API_Response {
	result := &API_Response{
		Name:     self.config.Name,
		Disabled: self.o.IsDisabled(),
		Restart:  self.o.IsRestart(),
		RunOnce:  self.o.IsRunOnce(),
		Secret:   self.config.Secret != "",
		CAS:      self.o.GetCAS(),
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
		result.KeyValue = self.o.GetKeyValue()
	}
	if !self.o.GetStarted().IsZero() {
		result.Started = &Timestamp{
			Time: self.o.GetStarted(),
			f:    self.patrol.config.Timestamp,
		}
	}
	if !self.o.GetLastSeen().IsZero() {
		result.LastSeen = &Timestamp{
			Time: self.o.GetLastSeen(),
			f:    self.patrol.config.Timestamp,
		}
	}
	return result
}
