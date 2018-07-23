package patrol

import (
	"time"
)

func (self *App) apiRequest(
	request *API_Request,
) bool {
	now := time.Now()
	// CAS / Non-Ping Attributes:
	//
	// we have to update our CAS BEFORE we update `lastseen, started, pid` and other ping values
	// these and other similar variables will modify our CAS!!!
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
	// Non-CAS / Ping Attributes:
	//
	// we need to compare our PID to our previous PID
	// if our PID does not match, we need to close our previous App and create a new App
	// if PID exists we're assumed to be in a Ping, so we can call triggers
	// PID is the only attribute that is required to be sent with a Ping
	// the reason for this is that in the future this is the only actions that WILL NOT require a correct CAS!!!
	if request.PID > 0 {
		// request PID exists
		// PID requires a ping, so all ping triggers are safe
		if self.o.GetPID() > 0 {
			// App PID exists
			if request.PID != self.o.GetPID() {
				// App PID does not match
				// this is a new App
				// close previous App
				self.close()
				// start a new App
				self.o.SetStarted(now)
				// set PID
				self.o.SetPID(request.PID)
				// call trigger
				if self.config.TriggerStartedPinged != nil {
					// we're going to unlock and call our trigger
					self.o.Unlock()
					self.config.TriggerStartedPinged(self)
					self.o.Lock()
				}
			} else {
				// PID matches
				// set lastseen
				self.o.SetLastSeen(now)
				// call trigger
				if self.config.TriggerPinged != nil {
					// we're going to unlock and call our trigger
					self.o.Unlock()
					self.config.TriggerPinged(self)
					self.o.Lock()
				}
			}
		} else {
			// App PID does not exist
			// set PID
			self.o.SetPID(request.PID)
			if self.o.GetStarted().IsZero() {
				// this is a new App
				self.o.SetStarted(now)
				// call trigger
				if self.config.TriggerStartedPinged != nil {
					// we're going to unlock and call our trigger
					self.o.Unlock()
					self.config.TriggerStartedPinged(self)
					self.o.Lock()
				}
			} else {
				// app was previously started
				// set lastseen
				self.o.SetLastSeen(now)
				// call trigger
				if self.config.TriggerPinged != nil {
					// we're going to unlock and call our trigger
					self.o.Unlock()
					self.config.TriggerPinged(self)
					self.o.Lock()
				}
			}
		}
	} else {
		// request PID doesn't exist
		if request.Ping {
			// set lastseen
			self.o.SetLastSeen(now)
			// call trigger
			if self.config.TriggerPinged != nil {
				// we're going to unlock and call our trigger
				self.o.Unlock()
				self.config.TriggerPinged(self)
				self.o.Lock()
			}
		}
	}
	return cas_valid
}
func (self *App) toggle(
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
func (self *App) Snapshot() *API_Response {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.apiResponse(api_endpoint_snapshot)
}
func (self *App) apiResponse(
	endpoint uint8,
) *API_Response {
	result := &API_Response{
		Name:     self.config.Name,
		PID:      self.o.GetPID(),
		Disabled: self.o.IsDisabled(),
		Restart:  self.o.IsRestart(),
		RunOnce:  self.o.IsRunOnce(),
		Secret:   self.config.Secret != "",
		CAS:      self.o.GetCAS(),
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
		result.KeyValue = self.o.GetKeyValue()
	}
	if !self.o.GetStarted().IsZero() {
		result.Started = &Timestamp{
			Time:            self.o.GetStarted(),
			TimestampFormat: self.patrol.config.Timestamp,
		}
	}
	if self.o.GetLastSeen().IsZero() {
		if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
			// if our app was running lastseen should exist
			if !self.o.GetStarted().IsZero() {
				// we should set lastseen to now
				// we're responsible for this service to always be running
				result.LastSeen = &Timestamp{
					Time:            time.Now(),
					TimestampFormat: self.patrol.config.Timestamp,
				}
			}
		}
	} else {
		result.LastSeen = &Timestamp{
			Time:            self.o.GetLastSeen(),
			TimestampFormat: self.patrol.config.Timestamp,
		}
	}
	return result
}
