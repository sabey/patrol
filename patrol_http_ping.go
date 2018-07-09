package patrol

import (
	"net/http"
)

func (self *Patrol) ServeHTTPPing(
	w http.ResponseWriter,
	r *http.Request,
) {
	// only App supports Ping
	self.serveHTTP(true, w, r)
}
