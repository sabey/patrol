package patrol

import (
	"encoding/json"
)

const (
	API_TOGGLE_STATE_ENABLE = iota + 1
	API_TOGGLE_STATE_DISABLE
	API_TOGGLE_STATE_RESTART
	API_TOGGLE_STATE_RUNONCE_ENABLE         // enable runonce
	API_TOGGLE_STATE_RUNONCE_DISABLE        // disable runonce
	API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE  // enable AND enable runonce
	API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE // enable AND disable runonce
)

const (
	LISTEN_HTTP_PORT_DEFAULT = 8421
	LISTEN_UDP_PORT_DEFAULT  = 1248
)

// Requests by Default are STATELESS - If no values are set then nothing is modified!
// The reason we're stateless by default is so that our UDP endpoint can make requests as if it were a HTTP GET Request
// UDP has the downside that if an error occurs a response will not be sent in return
type API_Request struct {
	// Unique Identifier
	ID string `json:"id,omitempty"`
	// Group: `app` or `service`
	Group string `json:"group,omitempty"`
	// Ping?
	// Only supported by either APP_KEEPALIVE_HTTP or APP_KEEPALIVE_UDP
	// If APP_KEEPALIVE_HTTP is used, the HTTP endpoint MUST be used
	// If APP_KEEPALIVE_UDP is used, the UDP endpoint MUST be used
	Ping bool `json:"ping,omitempty"`
	// App Process ID
	// Ping MUST be true if we wish to send a PID
	PID uint32 `json:"pid,omitempty"`
	// Toggle State
	//
	// API_TOGGLE_STATE_ENABLE = 1
	// API_TOGGLE_STATE_DISABLE = 2
	// API_TOGGLE_STATE_RESTART = 3
	// API_TOGGLE_STATE_RUNONCE_ENABLE = 4
	// API_TOGGLE_STATE_RUNONCE_DISABLE = 5
	// API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE = 6
	// API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE = 7
	//
	// API_TOGGLE_STATE_ENABLE: Enable App or Service
	// API_TOGGLE_STATE_DISABLE: Disable App or Service
	// API_TOGGLE_STATE_RESTART: Restart App or Service, Enable App or Service if Disabled
	// API_TOGGLE_STATE_RUNONCE_ENABLE: Enable RunOnce for App or Service
	// API_TOGGLE_STATE_RUNONCE_DISABLE: Disable RunOnce for App or Service
	// API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE: Enable App or Service and Enable RunOnce
	// API_TOGGLE_STATE_ENABLE_RUNONCE_DISABLE: Enable App or Service and Disable RunOnce
	Toggle uint8 `json:"toggle,omitempty"`
	// Return History?
	History bool `json:"history,omitempty"`
	// KeyValue
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
	// If KeyValueReplace is true, previous KeyValue will be replaced with KeyValue
	KeyValueReplace bool `json:"keyvalue-replace,omitempty"`
	// Secret is required to access the /api GET and POST endpoints
	Secret string `json:"secret,omitempty"`
	// CAS IS OPTIONAL
	// if CAS is NOT set: we will ignore it and we will override all of our values and state!!!
	// if CAS IS SET: we will only override values if our CAS is correct!
	// HOWEVER, we will ALWAYS update our PING/LastSeen value REGARDLESS OF CAS!!!
	// updating `Ping, LastSeen, or PID` will cause our CAS to be incremented!!!
	CAS uint64 `json:"cas,omitempty"`
}

func (self *API_Request) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}

// An API Response references our STATE at the time of Request
// If any values change or CAS is incremented, they will STILL reference the premodification state!
//
// When using UDP, we won't be able to respond with all of our data, we're going to have to limit our response size
// We're going to limit our response to: `id, group, pid, started, lastseen, disabled, restart, run-once, shutdown`
// We'll have to ignore `history, keyvalue, and errors`, if they're needed the HTTP endpoint should be used instead
type API_Response struct {
	// Unique Identifier
	ID string `json:"id,omitempty"`
	// Instance ID - UUIDv4 - Only exists IF we're running!
	InstanceID string `json:"instance-id,omitempty"`
	// Group: `app` or `service`
	Group string `json:"group,omitempty"`
	// Display Name
	Name string `json:"name,omitempty"`
	// App Process ID
	PID uint32 `json:"pid,omitempty"`
	// Timestamp App or Service started at
	Started *Timestamp `json:"started,omitempty"`
	// Timestamp App or Service was last seen
	LastSeen *Timestamp `json:"lastseen,omitempty"`
	// Is our App or Service Disabled?
	Disabled bool `json:"disabled,omitempty"`
	// Is our App or Service in a Restart state?
	Restart bool `json:"restart,omitempty"`
	// Is our App or Service set to RunOnce?
	RunOnce bool `json:"run-once,omitempty"`
	// Is Patrol in a Shutdown state?
	Shutdown bool `json:"shutdown,omitempty"`
	// History of previous App or Service states at the time of close()
	History []*History `json:"history,omitempty"`
	// Current state's KeyValue
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
	// Does this App or Service require a Secret to modify?
	Secret bool `json:"secret,omitempty"`
	// Did any Errors occur?
	Errors []string `json:"errors,omitempty"`
	// like all of our other values, CAS is a snapshot of our PREVIOUS state
	// we are NEVER going to return our current CAS after modifying our current state or values
	// the reason for this is that if a modification request is successful, we know our CAS is CAS + 1
	// if we were to take a snapshot, update our object, then get our CAS ---
	// we could never actually verify what our current state or values are!!!
	// the reason for this has to do with triggers, we NEVER KNOW when we're going to unlock and/or execute triggers!!!
	// there are going to be very many scenarios where an API request is made and our CAS is updated more than once!!!
	// we're never in a scenario where we take a snapshot, update, and get our CAS WITHOUT UNLOCKING!!!
	// if we want to make a clean CAS, we should do a REQUEST without modifying anything(no ping), then do a secondary request without incrementing CAS!
	CAS uint64 `json:"cas,omitempty"`
	// CASInvalid is the only exception to data that references our previous snapshot
	// We need to know if our CAS was successful or not!
	// I prefer to have this as invalid and not valid as most requests without a CAS will be valid!
	CASInvalid bool `json:"cas-invalid,omitempty"`
	// this is for unmarshal only
	patrol *Patrol
}

// I'm not sure how else to override/fix our timestamp when using a custom Timestamp Parse in Unmarshal
// we're going to clone our struct except for []History
// we will temporarily use json.RawMessage and then we will create a History object for each result
// this really isn't that aesthetic but it works well!
type api_response struct {
	ID         string                 `json:"id,omitempty"`
	InstanceID string                 `json:"instance-id,omitempty"`
	Group      string                 `json:"group,omitempty"`
	Name       string                 `json:"name,omitempty"`
	PID        uint32                 `json:"pid,omitempty"`
	Started    *Timestamp             `json:"started,omitempty"`
	LastSeen   *Timestamp             `json:"lastseen,omitempty"`
	Disabled   bool                   `json:"disabled,omitempty"`
	Restart    bool                   `json:"restart,omitempty"`
	RunOnce    bool                   `json:"run-once,omitempty"`
	Shutdown   bool                   `json:"shutdown,omitempty"`
	History    []json.RawMessage      `json:"history,omitempty"`
	KeyValue   map[string]interface{} `json:"keyvalue,omitempty"`
	Secret     bool                   `json:"secret,omitempty"`
	Errors     []string               `json:"errors,omitempty"`
	CAS        uint64                 `json:"cas,omitempty"`
	CASInvalid bool                   `json:"cas-invalid,omitempty"`
}

func (self *API_Response) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *API_Response) UnmarshalJSON(
	data []byte,
) error {
	result := &api_response{
		Started:  self.NewAPITimestamp(),
		LastSeen: self.NewAPITimestamp(),
	}
	if err := json.Unmarshal(data, result); err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	// unmarshal history
	if l := len(result.History); l > 0 {
		self.History = make([]*History, 0, l)
		for i := 0; i < l; i++ {
			h := self.NewAPIHistory()
			if err := json.Unmarshal(result.History[i], h); err != nil {
				return err
			}
			self.History = append(self.History, h)
		}
	}
	// fix response
	self.ID = result.ID
	self.InstanceID = result.InstanceID
	self.Group = result.Group
	self.Name = result.Name
	self.PID = result.PID
	self.Started = result.Started
	self.LastSeen = result.LastSeen
	self.Disabled = result.Disabled
	self.Restart = result.Restart
	self.RunOnce = result.RunOnce
	self.Shutdown = result.Shutdown
	self.KeyValue = result.KeyValue
	self.Secret = result.Secret
	self.Errors = result.Errors
	self.CAS = result.CAS
	self.CASInvalid = result.CASInvalid
	return nil
}
func (self *API_Response) NewAPITimestamp() *Timestamp {
	if !self.patrol.IsValid() {
		return &Timestamp{}
	}
	return self.patrol.NewAPITimestamp()
}
func (self *API_Response) NewAPIHistory() *History {
	if !self.patrol.IsValid() {
		return &History{
			Started:  &Timestamp{},
			LastSeen: &Timestamp{},
			Stopped:  &Timestamp{},
		}
	}
	return self.patrol.NewAPIHistory()
}

// use this when our response needs a reference to patrol for custom timestamp unmarshaling
func (self *Patrol) NewAPIResponse() *API_Response {
	return &API_Response{
		patrol: self,
	}
}
func (self *Patrol) NewAPIHistory() *History {
	return &History{
		Started:  self.NewAPITimestamp(),
		LastSeen: self.NewAPITimestamp(),
		Stopped:  self.NewAPITimestamp(),
	}
}
func (self *Patrol) NewAPITimestamp() *Timestamp {
	return &Timestamp{
		TimestampFormat: self.config.Timestamp,
	}
}
