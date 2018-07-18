package http

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sabey.co/patrol"
	"strings"
)

func Logs(
	w http.ResponseWriter,
	r *http.Request,
	p *patrol.Patrol,
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
	if !strings.HasPrefix(r.URL.Path, "/logs/") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(500)
		fmt.Fprintln(w, "HTTP Endpoint MUST have `/logs/` prefix!!!")
		return
	}
	// Filesystem
	path := filepath.Clean(r.URL.Path[5:])
	if path == "." {
		path = ""
	}
	source := c.WorkingDirectory + "/" + c.LogDirectory + "/" + path
	// Check if we're a DIRECTORY
	f, err := os.Stat(source)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(404)
		fmt.Fprintln(w, "404 page not found")
		return
	}
	if f.IsDir() {
		// this is a Directory!!!
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fi, err := ioutil.ReadDir(source)
		if err != nil {
			w.WriteHeader(500)
			log.Printf("./patrol/http.Logs(): failed to ReadDir: \"%s\"\n", err)
			fmt.Fprintln(w, "500 internal error")
			return
		}
		fmt.Fprintln(w, "<ul>")
		if path == "" || path == "/" {
			fmt.Fprintln(w, "<li><a href=\"../\">..</a></li>")
		} else {
			fmt.Fprintf(
				w,
				"<li><a href=\"../?%s\">..</a></li>\n",
				template.HTMLEscapeString(q.Encode()),
			)
		}
		for _, info := range fi {
			if info.IsDir() {
				fmt.Fprintf(
					w,
					"<li><a href=\"%s/?%s\">%s/</a></li>\n",
					info.Name(),
					template.HTMLEscapeString(q.Encode()),
					template.HTMLEscapeString(info.Name()),
				)
			} else {
				fmt.Fprintf(
					w,
					"<li><a href=\"%s?%s\">%s</a></li>\n",
					info.Name(),
					template.HTMLEscapeString(q.Encode()),
					template.HTMLEscapeString(info.Name()),
				)
			}
		}
		fmt.Fprintln(w, "</ul>")
		return
	}
	// this is a FILE
	http.ServeFile(w, r, source)
}
