package patrol

import (
	"encoding/json"
	"net/http"
)

func (self *Patrol) ServeHTTPStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	bs, _ := json.MarshalIndent(self.GetStatus(), "", "\t")
	w.WriteHeader(200)
	w.Write(bs)
	w.Write([]byte("\n"))
}
