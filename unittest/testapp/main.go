package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sabey.co/patrol"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	start := time.Now()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	unittesting_server := false
	if os.Getenv(patrol.PATROL_ENV_UNITTEST_KEY) == patrol.PATROL_ENV_UNITTEST_VALUE {
		// unittesting!!!
		unittesting_server = true
	}
	// get wd
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("failed get working directory: %s\n", err)
		os.Exit(253)
		return
	}
	log.Printf("wd: \"%s\"\n", wd)
	// we need to get our app id
	id := os.Getenv(patrol.APP_ENV_APP_ID)
	if id == "" {
		log.Fatalf("Patrol ID NOT set! Set ENV %s=ID\n", patrol.APP_ENV_APP_ID)
		return
	}
	// parse keepalive from our environment variable
	ka := os.Getenv(patrol.APP_ENV_KEEPALIVE)
	if ka == "" {
		log.Fatalf("KeepAlive NOT set! Set ENV %s=KEEPALIVE\n", patrol.APP_ENV_KEEPALIVE)
		return
	}
	keepalive, err := strconv.ParseUint(ka, 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse KeepAlive!")
		return
	}
	// create a ping channel
	ping := make(chan struct{})
	// we want to know if we managed to read any response
	// we require at least ONE successful message to be read in response!
	ping_read := false
	ping_done := false
	var ping_mu sync.Mutex
	// log and fmt will be spread out throughout this app
	// the reason for this is to send half of the msgs to stdout and the other to stderr
	log.Println("hello, I am a test app. I will run until you signal me to stop!")
	fmt.Printf("the time is now: %s\n", start)
	go func() {
		<-time.After(time.Minute * 5)
		log.Fatalln("testapp ran for too long, killing process")
	}()
	if keepalive == patrol.APP_KEEPALIVE_PID_PATROL {
		log.Printf("KeepAlive: APP_KEEPALIVE_PID_PATROL - we will NOT write PID: %d to file!", os.Getpid())
	} else if keepalive == patrol.APP_KEEPALIVE_PID_APP {
		log.Println("KeepAlive: APP_KEEPALIVE_PID_APP")
		// open testapp.pid
		// truncate and open file, create file if it doesn't exist
		file, err := os.OpenFile("testapp.pid", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("failed to open testapp.pid: \"%s\"\n", err)
			return
		}
		// write our PID to testapp.pid
		// append newline as well, clients should be prepared to handle whitespace themselves
		if _, err := file.WriteString(fmt.Sprintf("%d\n", os.Getpid())); err != nil {
			file.Close()
			log.Fatalf("failed to write to file: \"%s\"\n", err)
			return
		}
		if err := file.Close(); err != nil {
			log.Fatalf("failed to close file: \"%s\"\n", err)
			return
		}
		fmt.Printf("I've written our PID: %d to testapp.pid\n", os.Getpid())
	} else if keepalive == patrol.APP_KEEPALIVE_HTTP {
		log.Println("KeepAlive: APP_KEEPALIVE_HTTP")
		listeners := []string{}
		// unmarshal
		if err := json.Unmarshal([]byte(os.Getenv(patrol.APP_ENV_LISTEN_HTTP)), &listeners); err != nil {
			log.Fatalf("failed to unmarshal listeners: \"%s\"\n", err)
			return
		}
		if len(listeners) == 0 {
			log.Fatalln("no http listeners exist")
			return
		}
		for _, l := range listeners {
			if l == "" {
				log.Fatalln("empty http listeners found!")
				return
			}
		}
		log.Printf("http listeners found: \"%s\" using: \"%s\"\n", listeners, listeners[0])
		// start pinger
		go func() {
			log.Println("http starting to ping in 2 seconds")
			// we're going to wait 2 seconds BEFORE we start pinging
			<-time.After(time.Second * 2)
			log.Println("http starting to ping")
			p := 1
			for {
				select {
				// we're going to ping every second
				case <-time.After(time.Second):
					// build POST body
					extra := ""
					if unittesting_server {
						// we're going to trigger run_once
						extra = fmt.Sprintf(",\"toggle\":%d", patrol.API_TOGGLE_STATE_RUNONCE_ENABLE)
					}
					request := fmt.Sprintf(`{"id":"%s","group":"app","ping":true,"pid":%d%s}`, id, os.Getpid(), extra)
					log.Printf("ping: %d `%s`\n", p, request)
					response, err := http.Post(
						fmt.Sprintf("http://%s/api/", listeners[0]),
						`application/json`,
						strings.NewReader(request),
					)
					if err != nil {
						s := fmt.Sprintf("http failed to POST: \"%s\" Err: \"%s\"\n", listeners[0], err)
						if unittesting_server {
							log.Println(s)
						} else {
							log.Fatalln(s)
						}
						return
					}
					if response.StatusCode != 200 {
						log.Fatalf("http failed to POST: \"%s\" StatusCode: %d != 200\n", listeners[0], response.StatusCode)
						return
					}
					// read body
					body, err := ioutil.ReadAll(response.Body)
					if err != nil {
						response.Body.Close()
						s := fmt.Sprintf("http failed to read response.Body: \"%s\" Err: \"%s\"\n", listeners[0], err)
						if unittesting_server {
							log.Println(s)
						} else {
							log.Fatalln(s)
						}
						return
					}
					response.Body.Close()
					// read something
					resp := &patrol.API_Response{}
					// unmarshal
					if err = json.Unmarshal(body, resp); err != nil ||
						!resp.IsValid() {
						// failed to unmarshal
						log.Fatalf("http failed to unmarshal JSON - Err: \"%s\"\n", err)
						return
					}
					ping_mu.Lock()
					ping_read = true // successfully read something!
					ping_mu.Unlock()
					log.Printf("pinged: %d `%s`\n", p, body)
					p++
				case <-ping:
					// don't read this value
					// we're done pinging
					log.Println("http done pinging")
					return
				}
			}
		}()
	} else if keepalive == patrol.APP_KEEPALIVE_UDP {
		log.Println("KeepAlive: APP_KEEPALIVE_UDP")
		listeners := []string{}
		// unmarshal
		if err := json.Unmarshal([]byte(os.Getenv(patrol.APP_ENV_LISTEN_UDP)), &listeners); err != nil {
			log.Fatalf("failed to unmarshal listeners: \"%s\"\n", err)
			return
		}
		if len(listeners) == 0 {
			log.Fatalln("no udp listeners exist")
			return
		}
		for _, l := range listeners {
			if l == "" {
				log.Fatalln("empty udp listeners found!")
				return
			}
		}
		log.Printf("udp listeners found: \"%s\" using: \"%s\"\n", listeners, listeners[0])
		// dial our udp listener
		d, err := net.ResolveUDPAddr("udp", listeners[0])
		if err != nil {
			log.Fatalf("failed to resolve udp listener: \"%s\"\n", err)
			return
		}
		// we don't need to supply a local address, we will take what we can get
		conn, err := net.DialUDP("udp", nil, d)
		if err != nil {
			log.Fatalf("failed to dial udp: \"%s\"\n", err)
			return
		}
		defer conn.Close()
		// start write pinger
		go func() {
			log.Println("udp starting to ping in 2 seconds")
			// we're going to wait 2 seconds BEFORE we start pinging
			<-time.After(time.Second * 2)
			log.Println("udp starting to ping")
			p := 1
			for {
				select {
				// unlike HTTP we're going to ping every HALF second
				// when using HTTP we should be guaranteed for one message to be delivered within 3 seconds
				// here we should up our attempts just in case, otherwise we may only send 2 packets
				case <-time.After(time.Millisecond * 500):
					// build JSON body
					extra := ""
					if unittesting_server {
						// we're going to trigger run_once
						extra = fmt.Sprintf(",\"toggle\":%d", patrol.API_TOGGLE_STATE_RUNONCE_ENABLE)
					}
					request := fmt.Sprintf(`{"id":"%s","group":"app","ping":true,"pid":%d%s}`, id, os.Getpid(), extra)
					log.Printf("ping: %d `%s`\n", p, request)
					_, err := conn.Write([]byte(request))
					if err != nil {
						s := fmt.Sprintf("udp failed to write: \"%s\" Err: \"%s\"\n", listeners[0], err)
						if unittesting_server {
							log.Println(s)
						} else {
							log.Fatalln(s)
						}
						return
					}
					// DO NOT READ A RESPONSE HERE!!!
					p++
				case <-ping:
					// don't read this value
					// we're done pinging
					log.Println("udp done pinging")
					return
				}
			}
		}()
		// start read pinger
		go func() {
			log.Println("udp starting to read")
			p := 1
			for {
				select {
				case <-ping:
					// don't read this value
					// we're done pinging
					log.Println("udp done pinging")
					return
				default:
					// read response?
					body := make([]byte, 2048)
					n, _, err := conn.ReadFromUDP(body)
					if err != nil {
						// we failed to read
						ping_mu.Lock()
						if ping_done {
							// we're done!!!
							// ignore this error
							ping_mu.Unlock()
							return
						}
						ping_mu.Unlock()
						s := fmt.Sprintf("udp failed to read - Err: \"%s\"\n", err)
						if unittesting_server {
							log.Println(s)
						} else {
							log.Fatalln(s)
						}
						return
					}
					// read something
					resp := &patrol.API_Response{}
					// unmarshal
					if err = json.Unmarshal(body[:n], resp); err != nil ||
						!resp.IsValid() {
						// failed to unmarshal
						log.Fatalf("http failed to unmarshal JSON - Err: \"%s\"\n", err)
						return
					}
					ping_mu.Lock()
					ping_read = true // successfully read something!
					ping_mu.Unlock()
					log.Printf("pinged: %d `%s`\n", p, body[:n])
					p++
				}
			}
		}()
	} else {
		log.Fatalln("Unknown KeepAlive Method!")
		return
	}
	// create an unbuffered channel to listen for signals
	signals := make(chan os.Signal)
	// send signal notifications to our variable `signals`
	signal.Notify(
		signals,
		// we're going to listen for all signals
		// we're however only interested in these signals:
		// // kill -1 PID
		// syscall.SIGHUP,
		// // kill -2 PID
		// // this is the same as ctrl+c from the CLI
		// syscall.SIGINT,
		// // kill -9 PID
		// syscall.SIGKILL,
		// // we're NOT going to quit on these! we're just going to log that we received them
		// // kill -10 PID
		// syscall.SIGUSR1,
		// // kill -12 PID
		// syscall.SIGUSR2,
	)
	fmt.Println("we're going to block and wait for a signal")
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
		done := false
		select {
		case <-time.After(time.Second * 10):
			// we haven't received any signals, lets let people know
			log.Println("still waiting for you to signal me!")
		case sig := <-signals:
			// we've received a signal, let's figure out what one it was!
			// we can use a switch statement to determine our signal type
			switch sig {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, closing!")
				done = true
				break
			case syscall.SIGINT:
				// terminate process
				// ctrl+c
				log.Println("Received SIGINT, stop pinging!")
				// this signal is currently not referenced anywhere in patrol
				// we're doing to use this as a signal to stop pinging
				ping_mu.Lock()
				if !ping_read {
					log.Fatalln("failed to receive any responses - dying!")
				}
				if !ping_done {
					// we've got multiple goroutines relying on this
					// we have to use close instead of sending a struct
					close(ping)
					ping_done = true
				}
				ping_mu.Unlock()
				// we're NOT going to mark this process as done
				// we're going to wait for a different signal to say we're done
			case syscall.SIGKILL:
				log.Println("Received SIGKILL, closing!")
				done = true
				break
			case syscall.SIGUSR1:
				// we're gracefully shutting down
				fmt.Println("Received SIGUSR1, exiting NOW!")
				os.Exit(10)
				return
			case syscall.SIGUSR2:
				// Patrol is shutting down
				log.Println("Received SIGUSR2, exiting NOW!")
				os.Exit(12)
				return
			case syscall.SIGTERM:
				log.Println("Received SIGTERM, parent process died, closing!")
				done = true
				break
			default:
				// don't try to handle this signal
				log.Printf("Received Unknown?: %v\n", sig)
			}
		}
		if done {
			// exit loop and quit
			log.Println("dying in 15 seconds!")
			go func() {
				<-time.After(time.Second * 15)
				log.Fatalln("cya")
			}()
			break
		}
	}
	fmt.Printf("we ran for %s\n", time.Now().Sub(start))
	log.Println("good bye!")
}
