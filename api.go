package patrol

const (
	API_TOGGLE_NOOP = iota
	API_TOGGLE_ENABLE
	API_TOGGLE_DISABLE
	API_TOGGLE_RUNONCE
	API_TOGGLE_RESTART
)
const (
	LISTEN_HTTP_PORT_DEFAULT = 8421
	LISTEN_UDP_PORT_DEFAULT  = 1248
)

type API_Request struct {
	ID              string                 `json:"id,omitempty"`
	Group           string                 `json:"group,omitempty"`
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

type API_Response struct {
	ID       string                 `json:"id,omitempty"`
	Group    string                 `json:"group,omitempty"`
	PID      uint32                 `json:"pid,omitempty"`
	Started  *Timestamp             `json:"started,omitempty"`
	LastSeen *Timestamp             `json:"lastseen,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
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
