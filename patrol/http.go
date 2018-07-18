package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sabey.co/patrol"
	patrol_http "sabey.co/patrol/http"
)

func httpListen() string {
	return fmt.Sprintf("127.0.0.1:%d", patrol.LISTEN_HTTP_PORT_DEFAULT)
}
func HTTP() {
	defer log.Println("./patrol.HTTP(): Stopped!!!")
	listen := p.GetConfig().HTTP.Listen
	log.Printf("./patrol.HTTP(): Listen: \"%s\"\n", listen)
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Printf("./patrol.HTTP(): Failed to Listen: \"%s\"\n", err)
		return
	}
	defer l.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/status/", p.ServeHTTPStatus)
	mux.HandleFunc("/api/", p.ServeHTTPAPI)
	mux.HandleFunc("/stdout/", stdout)
	mux.HandleFunc("/stderr/", stderr)
	mux.HandleFunc("/logs/", logs)
	mux.HandleFunc("/", index)
	go func() {
		if err := http.Serve(l, mux); err != nil {
			log.Printf("./patrol.HTTP(): Serve Error: \"%s\"\n", err)
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
	patrol_http.Index(w, r, p)
}
func stdout(
	w http.ResponseWriter,
	r *http.Request,
) {
	patrol_http.STDOut(w, r, p)
}
func stderr(
	w http.ResponseWriter,
	r *http.Request,
) {
	patrol_http.STDErr(w, r, p)
}
func logs(
	w http.ResponseWriter,
	r *http.Request,
) {
	patrol_http.Logs(w, r, p)
}
