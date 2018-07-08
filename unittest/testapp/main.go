package main

import (
	"fmt"
	"github.com/sabey/patrol"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	start := time.Now()
	// parse keepalive from our environment variable
	ka := os.Getenv(patrol.APP_ENV_KEEPALIVE)
	if ka == "" {
		log.Fatalf("KeepAlive NOT set! Set ENV %s=ID\n", patrol.APP_ENV_KEEPALIVE)
		return
	}
	keepalive, err := strconv.ParseUint(ka, 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse KeepAlive!")
		return
	}
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
		log.Fatalln("KeepAlive: APP_KEEPALIVE_HTTP - NOT IMPLEMENTED")
	} else if keepalive == patrol.APP_KEEPALIVE_UDP {
		log.Fatalln("KeepAlive: APP_KEEPALIVE_UDP - NOT IMPLEMENTED")
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
				log.Println("Received SIGINT, closing!")
				done = true
				break
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
				log.Println("Received SIGTERM, parent process died, dying in 15 seconds!")
				go func() {
					<-time.After(time.Second * 15)
					log.Fatalln("cya")
				}()
			default:
				// don't try to handle this signal
				log.Printf("Received Unknown?: %v\n", sig)
			}
		}
		if done {
			// exit loop and quit
			break
		}
	}
	fmt.Printf("we ran for %s\n", time.Now().Sub(start))
	log.Println("good bye!")
}
