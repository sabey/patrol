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
	PID      uint32                 `json:"pid,omitempty"`
	Started  *Timestamp             `json:"started,omitempty"`
	LastSeen *Timestamp             `json:"lastseen,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
	Restart  bool                   `json:"restart,omitempty"`
	RunOnce  bool                   `json:"run-once,omitempty"`
	Shutdown bool                   `json:"shutdown,omitempty"`
	History  []*History             `json:"history,omitempty"`
	KeyValue map[string]interface{} `json:"keyvalue,omitempty"`
	Errors   []string               `json:"errors,omitempty"`
}

func (self *API_Response) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
