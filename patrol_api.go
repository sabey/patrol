package patrol

import (
	"strings"
)

const (
	api_endpoint_none = iota
	api_endpoint_http
	api_endpoint_udp
	api_endpoint_status
	api_endpoint_snapshot
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
		// handle response
		a, ok := self.apps[request.ID]
		if !ok {
			return &API_Response{
				Errors: []string{
					"Unknown App",
				},
			}
		}
		// validate secret
		if a.config.Secret != "" &&
			a.config.Secret != request.Secret {
			return &API_Response{
				Errors: []string{
					"Secret Invalid",
				},
			}
		}
		// validate endpoint
		// validate ping
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
			// validate ping endpoint
			if endpoint != api_endpoint_none {
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
		// validate PID
		// PID is the only attribute that is required to be sent with a Ping
		if request.PID > 0 {
			if a.config.KeepAlive != APP_KEEPALIVE_HTTP &&
				a.config.KeepAlive != APP_KEEPALIVE_UDP {
				// unknown ping method
				return &API_Response{
					Errors: []string{
						"PID Not Supported",
					},
				}
			}
			if !request.Ping {
				// bad ping method
				return &API_Response{
					Errors: []string{
						"PID Requires Ping",
					},
				}
			}
		}
		a.o.Lock()
		// we need to process our response before we update our object
		// we're interested in returning our previous state, since we know what our new state will be
		response := a.apiResponse(endpoint)
		// handle request
		response.CASInvalid = !a.apiRequest(request)
		a.o.Unlock()
		return response
	} else if request.Group == "service" ||
		request.Group == "services" {
		// handle response
		s, ok := self.services[request.ID]
		if !ok {
			return &API_Response{
				Errors: []string{
					"Unknown Service",
				},
			}
		}
		// validate secret
		if s.config.Secret != "" &&
			s.config.Secret != request.Secret {
			return &API_Response{
				Errors: []string{
					"Secret Invalid",
				},
			}
		}
		s.o.Lock()
		// we need to process our response before we update our object
		// we're interested in returning our previous state, since we know what our new state will be
		response := s.apiResponse(endpoint)
		// handle request
		response.CASInvalid = !s.apiRequest(request)
		s.o.Unlock()
		return response
	}
	return &API_Response{
		Errors: []string{
			"Unknown Group",
		},
	}
}
