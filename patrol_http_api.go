package patrol

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (self *Patrol) ServeHTTPAPI(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != "POST" &&
		r.Method != "GET" {
		// unknown method
		w.WriteHeader(405)
		bs, _ := json.MarshalIndent(
			&API_Response{
				Errors: []string{
					"Invalid Method",
				},
			},
			"", "\t",
		)
		w.Write(bs)
		w.Write([]byte("\n"))
		return
	}
	request := &API_Request{}
	if r.Method == "GET" {
		// regular API GET - we're using a query string here
		// we only support returning a response, no modifications
		q := r.URL.Query()
		request.ID = q.Get("id")
		request.Group = q.Get("group")
		if toggle, _ := strconv.ParseUint(q.Get("toggle"), 10, 64); toggle > 0 {
			request.Toggle = uint8(toggle)
		}
		request.History = len(q["history"]) > 0
		request.Secret = q.Get("secret")
		if cas, _ := strconv.ParseUint(q.Get("cas"), 10, 64); cas > 0 {
			request.CAS = cas
		}
	} else {
		// POST
		// read request
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(400)
			bs, _ := json.MarshalIndent(
				&API_Response{
					Errors: []string{
						"Invalid Body",
					},
				},
				"", "\t",
			)
			w.Write(bs)
			w.Write([]byte("\n"))
			return
		}
		// unmarshal
		err = json.Unmarshal(body, request)
		if err != nil ||
			!request.IsValid() {
			w.WriteHeader(400)
			bs, _ := json.MarshalIndent(
				&API_Response{
					Errors: []string{
						"Invalid Request",
					},
				},
				"", "\t",
			)
			w.Write(bs)
			w.Write([]byte("\n"))
			return
		}
	}
	response := self.api(api_endpoint_http, request)
	if len(response.Errors) > 0 {
		w.WriteHeader(400)
	} else {
		w.WriteHeader(200)
	}
	bs, _ := json.MarshalIndent(response, "", "\t")
	w.Write(bs)
	w.Write([]byte("\n"))
}
