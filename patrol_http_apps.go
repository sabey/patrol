package patrol

import (
	"encoding/json"
	"net/http"
)

func (self *Patrol) ServeHTTPApps(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// this is just a basic array output
	apps := make([]string, 0, len(self.apps))
	for k, _ := range self.apps {
		apps = append(apps, k)
	}
	w.WriteHeader(200)
	bs, _ := json.Marshal(apps)
	w.Write(bs)
}
