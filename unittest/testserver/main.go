package main

import (
	"flag"
	"fmt"
	"github.com/sabey/patrol"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	config_path = flag.String("config", "config.json", "path to patrol config file")
)

func main() {
	start := time.Now()
	if !flag.Parsed() {
		flag.Parse()
	}
	// ideally in the future we may want to read from stdin
	// right now we can use different config files
	config, err := patrol.LoadConfig(*config_path)
	if err != nil {
		log.Printf("./unittest/testserver.main(): failed to Load Patrol Config: %s\n", err)
		os.Exit(254)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("./unittest/testserver.main(): failed get working directory: %s\n", err)
		os.Exit(253)
		return
	}
	shutdown := make(chan struct{})
	// create listeners
	need_http := false
	need_udp := false
	log.Println("./unittest/testserver.main(): Finding Listeners")
	for _, a := range config.Apps {
		if a.KeepAlive == patrol.APP_KEEPALIVE_HTTP {
			need_http = true
		} else if a.KeepAlive == patrol.APP_KEEPALIVE_UDP {
			need_udp = true
		}
		// prefix working directories!!!
		a.WorkingDirectory = filepath.Clean(wd + "/../" + a.WorkingDirectory)
	}
	http_listen := ""
	//udp_listen := ""
	if need_http {
		if config.HTTP != nil {
			http_listen = config.HTTP.Listen
		}
		if http_listen == "" {
			http_listen = fmt.Sprintf("127.0.0.1:%d", patrol.LISTEN_HTTP_PORT_DEFAULT)
		}
		// we have to append our address to patrol
		config.ListenHTTP = []string{http_listen}
	}
	// create patrol
	log.Println("./unittest/testserver.main(): Creating Patrol")
	p, err := patrol.CreatePatrol(config)
	if err != nil {
		log.Printf("./unittest/testserver.main(): failed to Create Patrol: %s\n", err)
		os.Exit(255)
		return
	}
	// create listenrs
	if need_http || need_udp {
		log.Println("./unittest/testserver.main(): Creating Listeners")
	}
	if need_http {
		log.Printf("./unittest/testserver.main(): Creating HTTP Listener: \"%s\"\n", http_listen)
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/status/", p.ServeHTTPStatus)
			mux.HandleFunc("/api/", p.ServeHTTPAPI)
			mux.HandleFunc("/ping/", p.ServeHTTPPing)
			mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				// The "/" pattern matches everything, so we need to check that we're at the root here.
				if req.URL.Path != "/" {
					http.NotFound(w, req)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(200)
				fmt.Fprintf(w, `Please use one of these endpoints:<br /><br />
GET /status/<br />
GET /api/?group=(app||service)&amp;id=app<br />
POST /api/<br />
POST /ping/
`)
			})
			http.ListenAndServe(http_listen, mux)
			// discard error
			// signal shutdown
			shutdown <- struct{}{}
		}()
	}
	// start patrol
	log.Println("./unittest/testserver.main(): Starting Patrol")
	p.Start()
	// create an unbuffered channel to listen for signals
	signals := make(chan os.Signal)
	// send signal notifications to our variable `signals`
	signal.Notify(
		signals,
		// notify of all signals
		/*syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGTERM,
		syscall.SIGTSTP,*/
	)
	done := false
	for {
		// https://en.wikipedia.org/wiki/Unix_signal
		/*
			jackson@Monolith:~$ kill -l
			1) SIGHUP		2) SIGINT		3) SIGQUIT		4) SIGILL		5) SIGTRAP
			6) SIGABRT		7) SIGBUS		8) SIGFPE		9) SIGKILL		10) SIGUSR1
			11) SIGSEGV		12) SIGUSR2		13) SIGPIPE		14) SIGALRM		15) SIGTERM
			16) SIGSTKFLT	17) SIGCHLD		18) SIGCONT		19) SIGSTOP		20) SIGTSTP
			21) SIGTTIN		22) SIGTTOU		23) SIGURG		24) SIGXCPU		25) SIGXFSZ
			26) SIGVTALRM	27) SIGPROF		28) SIGWINCH	29) SIGIO		30) SIGPWR
			31) SIGSYS		34) SIGRTMIN	35) SIGRTMIN+1	36) SIGRTMIN+2	37) SIGRTMIN+3
			38) SIGRTMIN+4	39) SIGRTMIN+5	40) SIGRTMIN+6	41) SIGRTMIN+7	42) SIGRTMIN+8
			43) SIGRTMIN+9	44) SIGRTMIN+10	45) SIGRTMIN+11	46) SIGRTMIN+12	47) SIGRTMIN+13
			48) SIGRTMIN+14	49) SIGRTMIN+15	50) SIGRTMAX-14	51) SIGRTMAX-13	52) SIGRTMAX-12
			53) SIGRTMAX-11	54) SIGRTMAX-10	55) SIGRTMAX-9	56) SIGRTMAX-8	57) SIGRTMAX-7
			58) SIGRTMAX-6	59) SIGRTMAX-5	60) SIGRTMAX-4	61) SIGRTMAX-3	62) SIGRTMAX-2
			63) SIGRTMAX-1	64) SIGRTMAX
		*/
		// handle signal and shutdown
		select {
		case <-shutdown:
			log.Println("./unittest/testserver.main(): listener shutdown!")
			done = true
			break
		case sig := <-signals:
			switch sig {
			case syscall.SIGHUP:
				// Hangup / ssh broken pipe
				log.Println("./unittest/testserver.main(): SIGHUP")
				done = true
				break
			case syscall.SIGINT:
				// terminate process
				// ctrl+c
				log.Println("./unittest/testserver.main(): SIGINT")
				done = true
				break
			case syscall.SIGQUIT:
				// ctrl+4 or ctrl+|
				log.Println("./unittest/testserver.main(): SIGQUIT")
				done = true
				break
			case syscall.SIGKILL:
				// kill -9
				// shutdown NOW
				log.Fatalf("./unittest/testserver.main(): SIGKILL")
				done = true
				break
			case syscall.SIGTERM:
				// killall service
				// gracefully shutdown NOW
				log.Println("./unittest/testserver.main(): SIGTERM")
				done = true
				break
			case syscall.SIGTSTP:
				// this will cause the program to go to the background if in a cli
				// ctrl+z
				done = true
				break
			case syscall.SIGUSR1:
				// unreserved signal - handle however we want
			case syscall.SIGUSR2:
				// unreserved signal - handle however we want
			default:
				// unknown
				// do nothing
				log.Printf("./unittest/testserver.main(): Unknown Signal Ignored: \"%v\"\n", sig)
			}
		}
		if done {
			break
		}
	}
	log.Println("./unittest/testserver.main(): Stopping Patrol")
	p.Stop()
	// wait for patrol to stop
	log.Println("./unittest/testserver.main(): Waiting for Patrol to stop!")
	for {
		// we're going to add a saftey measure incase we fail to stop
		go func() {
			<-time.After(time.Minute * 3)
			log.Fatalln("./unittest/testserver.main(): Failed to Stop Patrol, Dying!")
		}()
		if !p.IsRunning() {
			// we're done!
			break
		}
	}
	log.Printf("./unittest/testserver.main(): Patrol ran for: %s\n", time.Now().Sub(start))
	log.Println("good bye!")
}
