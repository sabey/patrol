package patrol

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

/*
request by default is stateless, if no values are set then nothing is modified
the reason we're stateless by default but share the same object is that we use this with UDP as well
a UDP request can send us a request with no changes and get a state back in response
our http endpoint in addition will REQUIRE us to use POST to make a modification
*/
type API_Request struct {
	ID              string                 `json:"id,omitempty"`
	Group           string                 `json:"group,omitempty"`
	Ping            bool                   `json:"ping,omitempty"`
	PID             uint32                 `json:"pid,omitempty"`
	Toggle          uint8                  `json:"toggle,omitempty"`
	History         bool                   `json:"history,omitempty"`
	KeyValue        map[string]interface{} `json:"keyvalue,omitempty"`
	KeyValueReplace bool                   `json:"keyvalue-replace,omitempty"`
	Secret          string                 `json:"secret,omitempty"`
	// CAS is OPTIONAL
	// if CAS is NOT set: we will ignore it and we will override all of our values and state!!!
	// if CAS IS SET: we will only override values if our CAS is correct!
	// HOWEVER, we will ALWAYS update our PING/lastseen value REGARDLESS OF CAS!!!
	// updating `Ping, LastSeen, or PID` will cause our CAS to be incremented!!!
	CAS uint64 `json:"cas,omitempty"`
}

func (self *API_Request) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}

/*
when using UDP, we won't be able to respond with all of our data, we're going to have to limit our response size
we're going to limit our response to: `id, group, pid, started, lastseen, disabled, restart, run-once, shutdown`
we'll have to ignore `history, keyvalue, and errors`, if they're needed the JSON endpoint should be used instead
*/
type API_Response struct {
	ID       string                 `json:"id,omitempty"`
	Group    string                 `json:"group,omitempty"`
	Name     string                 `json:"name,omitempty"`
	PID      uint32                 `json:"pid,omitempty"`
	Started  *Timestamp             `json:"started,omitempty"`
	LastSeen *Timestamp             `json:"lastseen,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
	Restart  bool                   `json:"restart,omitempty"`
	RunOnce  bool                   `json:"run-once,omitempty"`
	Shutdown bool                   `json:"shutdown,omitempty"`
	History  []*History             `json:"history,omitempty"`
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
	Secret   bool                   `json:"secret,omitempty"`
	Errors   []string               `json:"errors,omitempty"`
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
	// cas-invalid is the only exception to data that references our previous snapshot
	// we need to know if our CAS was successful or not!
	// I prefer to have this as invalid not valid as most requests without a CAS will be valid!
	CASInvalid bool `json:"cas-invalid,omitempty"`
}

func (self *API_Response) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
