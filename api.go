package patrol

const (
	API_TOGGLE_NOOP = iota
	API_TOGGLE_ENABLE
	API_TOGGLE_DISABLE
	API_TOGGLE_RUNONCE
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
	Started  PatrolTimestamp        `json:"started,omitempty"`
	LastSeen PatrolTimestamp        `json:"lastseen,omitempty"`
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
