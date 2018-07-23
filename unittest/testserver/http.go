package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sabey.co/patrol"
)

func httpListen() string {
	return fmt.Sprintf("127.0.0.1:%d", patrol.LISTEN_HTTP_PORT_DEFAULT)
}
func HTTP() {
	defer log.Println("./patrol/unittest/testserver.HTTP(): Stopped!!!")
	listen := p.GetConfig().HTTP.Listen
	log.Printf("./patrol/unittest/testserver.HTTP(): Listen: \"%s\"\n", listen)
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Printf("./patrol/unittest/testserver.HTTP(): Failed to Listen: \"%s\"\n", err)
		return
	}
	defer l.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/status/", p.ServeHTTPStatus)
	mux.HandleFunc("/api/", p.ServeHTTPAPI)
	mux.HandleFunc("/", index)
	go func() {
		if err := http.Serve(l, mux); err != nil {
			log.Printf("./patrol/unittest/testserver.HTTP(): Serve Error: \"%s\"\n", err)
		}
		// call shutdown
		shutdown()
	}()
	// wait for shutdown
	<-shutdown_c
}
func index(
	w http.ResponseWriter,
	r *http.Request,
) {
	// The "/" pattern matches everything, so we need to check that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, `Please use one of these endpoints:<br /><br />
GET /status/<br />
GET /api/?group=(app||service)&amp;id=app<br />
POST /api/
`)
}
