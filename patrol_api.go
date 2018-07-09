package patrol

import (
	"strings"
	"time"
)

func (self *Patrol) API(
	request *API_Request,
) *API_Response {
	// API supports both App and Service
	return self.api(false, request)
}
func (self *Patrol) api(
	ping bool,
	request *API_Request,
) *API_Response {
	if !request.IsValid() {
		return &API_Response{
			Errors: []string{
				"Request NIL",
			},
		}
	}
	if ping && (request.Group == "service" ||
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
		a.mu.Lock()
		defer a.mu.Unlock()
		// we need to process our response before we update our object
		// we're interested in returning our previous state, since we know what our new state will be
		response := a.apiResponse()
		// handle response
		if ping &&
			(a.config.KeepAlive == APP_KEEPALIVE_HTTP ||
				a.config.KeepAlive == APP_KEEPALIVE_UDP) {
			// only HTTP and UDP can update lastseen by API
			a.lastseen = time.Now()
		}
		a.apiRequest(ping, request)
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
		response := s.apiResponse()
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
