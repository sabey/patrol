package patrol

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func (self *Patrol) serveHTTP(
	ping bool,
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if (ping && r.Method != "POST") ||
		(!ping && !(r.Method == "POST" ||
			r.Method == "GET")) {
		w.WriteHeader(405)
		bs, _ := json.Marshal(
			&API_Response{
				Errors: []string{
					"Invalid Method",
				},
			},
		)
		w.Write(bs)
		return
	}
	request := &API_Request{}
	if r.Method == "GET" {
		// regular API Get
		// we only support returning a response, no modifications
		q := r.URL.Query()
		request.ID = q.Get("id")
		request.Group = q.Get("group")
		request.History = len(q["history"]) > 0
	} else if r.Method == "POST" {
		// read request
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(400)
			bs, _ := json.Marshal(
				&API_Response{
					Errors: []string{
						"Invalid Body",
					},
				},
			)
			w.Write(bs)
			return
		}
		// unmarshal
		err = json.Unmarshal(body, request)
		if err != nil ||
			!request.IsValid() {
			w.WriteHeader(400)
			bs, _ := json.Marshal(
				&API_Response{
					Errors: []string{
						"Invalid Request",
					},
				},
			)
			w.Write(bs)
			return
		}
	}
	var response *API_Response
	if ping {
		response = self.Ping(request)
	} else {
		response = self.Ping(request)
	}
	if len(response.Errors) > 0 {
		w.WriteHeader(400)
	} else {
		w.WriteHeader(200)
	}
	bs, _ := json.Marshal(response)
	w.Write(bs)
}
