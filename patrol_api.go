package patrol

import (
	"strings"
	"time"
)

const (
	api_endpoint_none = iota
	api_endpoint_http
	api_endpoint_udp
)

func (self *Patrol) API(
	request *API_Request,
) *API_Response {
	// do not validate endpoint
	return self.api(
		api_endpoint_none,
		request,
	)
}
func (self *Patrol) api(
	endpoint uint8,
	request *API_Request,
) *API_Response {
	if !request.IsValid() {
		return &API_Response{
			Errors: []string{
				"Request NIL",
			},
		}
	}
	if request.Ping && (request.Group == "service" ||
		request.Group == "services") {
		// services don't support Ping currently
		return &API_Response{
			Errors: []string{
				"Services don't support Ping",
			},
		}
	}
	request.ID = strings.ToLower(request.ID)
	request.Group = strings.ToLower(request.Group)
	if request.Group == "app" ||
		request.Group == "apps" {
		a, ok := self.apps[request.ID]
		if !ok {
			return &API_Response{
				Errors: []string{
					"Unknown App",
				},
			}
		}
		// validate endpoint
		if request.Ping {
			if a.config.KeepAlive != APP_KEEPALIVE_HTTP &&
				a.config.KeepAlive != APP_KEEPALIVE_UDP {
				// unknown ping method
				return &API_Response{
					Errors: []string{
						"Ping Not Supported",
					},
				}
			}
			if endpoint != api_endpoint_none {
				// validate ping endpoint
				if (a.config.KeepAlive == APP_KEEPALIVE_HTTP && endpoint != api_endpoint_http) ||
					(a.config.KeepAlive == APP_KEEPALIVE_UDP && endpoint != api_endpoint_udp) {
					return &API_Response{
						Errors: []string{
							"Invalid Ping Endpoint",
						},
					}
				}
			}
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		// we need to process our response before we update our object
		// we're interested in returning our previous state, since we know what our new state will be
		response := a.apiResponse(false)
		// handle response
		if request.Ping {
			// only HTTP and UDP can update lastseen by API
			a.lastseen = time.Now()
			// we need to check if we ever started this app
			// when we initially load patrol, there's a chance we could have APp that are STILL running
			// if they ping us and the App was never started, we have to set started and we won't restart this App!
			//
			// we don't want to do multiple triggers here
			// one is good enough, either way it will both represent a ping
			// but started ping will represent started from a ping
			triggered := false
			if a.started.IsZero() {
				// app was previously started
				a.started = a.lastseen
				// call ping started trigger
				if a.config.TriggerStartedPinged != nil {
					// use goroutine to avoid deadlock
					triggered = true
					go a.config.TriggerStartedPinged(a)
				}
			}
			// call ping trigger
			if !triggered && a.config.TriggerPinged != nil {
				// use goroutine to avoid deadlock
				go a.config.TriggerPinged(a)
			}
		}
		a.apiRequest(request)
		return response
	} else if request.Group == "service" ||
		request.Group == "services" {
		s, ok := self.services[request.ID]
		if !ok {
			return &API_Response{
				Errors: []string{
					"Unknown Service",
				},
			}
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		// we need to process our response before we update our object
		// we're interested in returning our previous state, since we know what our new state will be
		response := s.apiResponse(false)
		// handle response
		// we can't ever update service lastseen by ping or by api
		s.apiRequest(request)
		return response
	}
	return &API_Response{
		Errors: []string{
			"Unknown Group",
		},
	}
}
