package patrol

import (
	"net/http"
)

func (self *Patrol) ServeHTTPAPI(
	w http.ResponseWriter,
	r *http.Request,
) {
	self.serveHTTP(false, w, r)
}
