package patrol

import (
	"time"
)

func (self *App) apiRequest(
	request *API_Request,
) {
	now := time.Now()
	// we need to compare our PID to our previous PID
	// if our PID does not match, we need to close our previous App and create a new App
	// if PID exists we're assumed to be in a Ping, so we can call triggers
	// PID is the only attribute that is required to be sent with a Ping
	// the reason for this is that in the future this is the only actions that WILL NOT require a correct CAS!!!
	if request.PID > 0 {
		// request PID exists
		// PID requires a ping, so all ping triggers are safe
		if self.pid > 0 {
			// App PID exists
			if request.PID != self.pid {
				// App PID does not match
				// this is a new App
				// close previous App
				self.close()
				// start a new App
				self.started = now
				// set PID
				self.pid = request.PID
				// call trigger
				if self.config.TriggerStartedPinged != nil {
					// we're going to unlock and call our trigger
					self.mu.Unlock()
					self.config.TriggerStartedPinged(self)
					self.mu.Lock()
				}
			} else {
				// PID matches
				// set lastseen
				self.lastseen = now
				// call trigger
				if self.config.TriggerPinged != nil {
					// we're going to unlock and call our trigger
					self.mu.Unlock()
					self.config.TriggerPinged(self)
					self.mu.Lock()
				}
			}
		} else {
			// App PID does not exist
			// set PID
			self.pid = request.PID
			if self.started.IsZero() {
				// this is a new App
				self.started = now
				// call trigger
				if self.config.TriggerStartedPinged != nil {
					// we're going to unlock and call our trigger
					self.mu.Unlock()
					self.config.TriggerStartedPinged(self)
					self.mu.Lock()
				}
			} else {
				// app was previously started
				// set lastseen
				self.lastseen = now
				// call trigger
				if self.config.TriggerPinged != nil {
					// we're going to unlock and call our trigger
					self.mu.Unlock()
					self.config.TriggerPinged(self)
					self.mu.Lock()
				}
			}
		}
	} else {
		// request PID doesn't exist
		if request.Ping {
			// set lastseen
			self.lastseen = now
			// call trigger
			if self.config.TriggerPinged != nil {
				// we're going to unlock and call our trigger
				self.mu.Unlock()
				self.config.TriggerPinged(self)
				self.mu.Lock()
			}
		}
	}
	// non ping attributes:
	// TODO: add optional CAS
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
func (self *App) toggle(
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
func (self *App) Snapshot() *API_Response {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.apiResponse(api_endpoint_snapshot)
}
func (self *App) apiResponse(
	endpoint uint8,
) *API_Response {
	result := &API_Response{
		PID:      self.pid,
		Disabled: self.disabled,
		Restart:  self.restart,
		RunOnce:  self.run_once,
	}
	if endpoint != api_endpoint_status {
		// we don't need these values for individual status objects
		result.ID = self.id
		result.Group = "app"
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
