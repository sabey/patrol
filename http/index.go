package http

import (
	"fmt"
	"github.com/sabey/patrol"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	wd string
)

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		log.Printf("./patrol/http.init(): failed to get Working Directory: %s\n", err)
		os.Exit(250)
		return
	}
}
func Index(
	w http.ResponseWriter,
	r *http.Request,
	p *patrol.Patrol,
) {
	if r.URL.Path == "/" {
		// Index
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var post *StatusPost
		if r.Method == "POST" {
			// parse POST
			post = &StatusPost{}
			if err := r.ParseForm(); err == nil {
				post.ID = r.PostFormValue("id")
				post.Group = r.PostFormValue("group")
				post.Secret = r.PostFormValue("secret")
				if r.PostFormValue("disable") != "" {
					post.Toggle = patrol.API_TOGGLE_STATE_DISABLE
				} else if r.PostFormValue("enable") != "" {
					post.Toggle = patrol.API_TOGGLE_STATE_ENABLE
				} else if r.PostFormValue("restart") != "" {
					post.Toggle = patrol.API_TOGGLE_STATE_RESTART
				} else if r.PostFormValue("runonce-enable") != "" {
					post.Toggle = patrol.API_TOGGLE_STATE_RUNONCE_ENABLE
				} else if r.PostFormValue("runonce-disable") != "" {
					post.Toggle = patrol.API_TOGGLE_STATE_RUNONCE_DISABLE
				}
				if post.Toggle > 0 {
					if post.Group == "app" {
						if app := p.GetApp(post.ID); app.IsValid() {
							c := app.GetConfig()
							// validate secret
							if c.Secret == "" ||
								c.Secret == post.Secret {
								// toggle
								app.Toggle(post.Toggle)
								// success!
								w.Header().Set("Location", "/")
								w.WriteHeader(302)
								return
							} else {
								post.Error = "Secret Invalid"
							}
						} else {
							post.Error = "Unknown App"
						}
					} else if post.Group == "service" {
						if service := p.GetService(post.ID); service.IsValid() {
							c := service.GetConfig()
							// validate secret
							if c.Secret == "" ||
								c.Secret == post.Secret {
								// toggle
								service.Toggle(post.Toggle)
								// success!
								w.Header().Set("Location", "/")
								w.WriteHeader(302)
								return
							} else {
								post.Error = "Secret Invalid"
							}
						} else {
							post.Error = "Unknown Service"
						}
					} else {
						post.Error = "Unknown Group"
					}
				} else {
					post.Error = "Toggle Invalid"
				}
			} else {
				post.Error = "Bad POST"
			}
		}
		// we do NOT care about the overhead of parsing our template on each request
		// we might change this later, but if we do not have a lot of traffic there's no reason to
		t, err := template.ParseGlob("tmpl/*.tmpl")
		if err != nil {
			log.Printf("/patrol/http.Index(): failed to parse templates: \"%s\"\n", err)
			w.WriteHeader(500)
			// we do not care about printing errors
			// only admins should have access to this endpoint
			fmt.Fprintf(
				w,
				"<pre>"+
					template.HTMLEscapeString(
						fmt.Sprintf(
							"Failed to Parse \"index.tmpl\"\nError: \"%s\"\n",
							err.Error(),
						),
					)+
					"</pre>\n",
			)
			return
		}
		data := &Data{
			Patrol:     p,
			Status:     p.GetStatus(),
			StatusPost: post,
			Now:        time.Now(),
		}
		if err := t.ExecuteTemplate(w, "index.tmpl", data); err != nil {
			log.Printf("/patrol/http.Index(): failed to execute template: \"%s\"\n", err)
			// successfully writing to our response writer will cause us to issue a 200
			// if an error occurs, either there is an issue with our template or nothing has been written
			// we can now write 500
			w.WriteHeader(500)
			// chances are if something was written it will be a mixed response
			fmt.Fprintf(
				w,
				"<pre>"+
					template.HTMLEscapeString(
						fmt.Sprintf(
							"Failed to Execute \"index.tmpl\"\nError: \"%s\"\n",
							err.Error(),
						),
					)+
					"</pre>\n",
			)
		}
		return
	}
	// Filesystem
	path := filepath.Clean(r.URL.Path)
	if path == "." {
		path = ""
	}
	http.ServeFile(w, r, wd+"/www/"+path)
}

type Data struct {
	Patrol     *patrol.Patrol     `json:"patrol,omitempty"`
	Status     *patrol.API_Status `json:"status,omitempty"`
	StatusPost *StatusPost        `json:"status-post,omitempty"`
	Now        time.Time          `json:"now,omitempty"`
	X          interface{}        `json:"x,omitempty"`
}
type StatusPost struct {
	ID     string `json:"id,omitempty"`
	Group  string `json:"group,omitempty"`
	Secret string `json:"secret,omitempty"`
	Toggle uint8  `json:"toggle,omitempty"`
	Error  string `json:"error,omitempty"`
}
