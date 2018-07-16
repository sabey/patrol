package patrol

type API_Status struct {
	Apps     map[string]*API_Response `json:"apps,omitempty"`
	Services map[string]*API_Response `json:"service,omitempty"`
	Started  *Timestamp               `json:"started,omitempty"`
	Shutdown bool                     `json:"shutdown,omitempty"`
}

func (self *Patrol) getStatus() *API_Status {
	self.mu.RLock()
	started := self.ticker_running
	shutdown := self.shutdown
	self.mu.RUnlock()
	result := &API_Status{
		Apps:     make(map[string]*API_Response),
		Services: make(map[string]*API_Response),
		Shutdown: shutdown,
	}
	if !started.IsZero() {
		result.Started = &Timestamp{
			Time: started,
			f:    self.config.Timestamp,
		}
	}
	for id, app := range self.apps {
		app.mu.RLock()
		result.Apps[id] = app.apiResponse(api_endpoint_status)
		app.mu.RUnlock()
	}
	for id, service := range self.services {
		service.mu.RLock()
		result.Services[id] = service.apiResponse(api_endpoint_status)
		service.mu.RUnlock()
	}
	return result
}
