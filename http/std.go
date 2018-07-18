package http

import (
	"fmt"
	"net/http"
	"sabey.co/patrol"
	"strconv"
)

func STDOut(
	w http.ResponseWriter,
	r *http.Request,
	p *patrol.Patrol,
) {
	std(w, r, p, true)
}
func STDErr(
	w http.ResponseWriter,
	r *http.Request,
	p *patrol.Patrol,
) {
	std(w, r, p, false)
}
func std(
	w http.ResponseWriter,
	r *http.Request,
	p *patrol.Patrol,
	out bool,
) {
	q := r.URL.Query()
	group := q.Get("group")
	if group != "app" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintln(w, "Unknown Group")
		return
	}
	id := q.Get("id")
	app := p.GetApp(id)
	if !app.IsValid() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintln(w, "Unknown App")
		return
	}
	secret := q.Get("secret")
	// get config
	c := app.GetConfig()
	// validate secret
	if c.Secret != "" &&
		c.Secret != secret {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintln(w, "Secret Invalid")
		return
	}
	var last uint64 = 0
	if l := q.Get("last"); l != "" {
		var err error
		last, err = strconv.ParseUint(l, 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(405)
			fmt.Fprintln(w, "?last=INVALID")
			return
		}
		if last > 1024 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(405)
			fmt.Fprintln(w, "?last= does not support over 1024 lines")
			return
		}
		// I'm not sure what I'll do with this function yet
		// I don't currently reference it in our GUI yet
		// this should be support but I'm not going to work on it just yet
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(405)
		fmt.Fprintln(w, "?last= NOT supported yet!")
		return
	}
	path := ""
	if out {
		path = app.GetStdoutLog()
	} else {
		path = app.GetStderrLog()
	}
	if path == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(404)
		fmt.Fprintln(w, "404 log type doesn't exist")
		return
	}
	if last == 0 {
		// we're going to print our entire file
		http.ServeFile(w, r, path)
		return
	}
	// we need to seek our file and search for our last x lines
}
