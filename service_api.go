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
func (self *Service) apiResponse() *API_Response {
	result := &API_Response{
		ID:    self.id,
		Group: "service",
		Started: PatrolTimestamp{
			Time: self.started,
			f:    self.patrol.config.Timestamp,
		},
		LastSeen: PatrolTimestamp{
			Time: self.lastseen,
			f:    self.patrol.config.Timestamp,
		},
		Disabled: self.disabled,
		Shutdown: self.patrol.shutdown,
		KeyValue: self.getKeyValue(),
	}
	return result
}
