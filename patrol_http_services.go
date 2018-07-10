package patrol

import (
	"encoding/json"
	"net/http"
)

func (self *Patrol) ServeHTTPServices(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// this is just a basic array output
	services := make([]string, 0, len(self.services))
	for k, _ := range self.services {
		services = append(services, k)
	}
	w.WriteHeader(200)
	bs, _ := json.Marshal(services)
	w.Write(bs)
}
