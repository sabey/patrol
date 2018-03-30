package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	keepalive = flag.Int("keepalive", 2, "KeepAlive Method: (1: \"Patrol\", 2: \"App\", 3: \"HTTP\", 4: \"UDP\")")
)

func main() {
	start := time.Now()
	if !flag.Parsed() {
		flag.Parse()
	}
	// log and fmt will be spread out throughout this app
	// the reason for this is to send half of the msgs to stdout and the other to stderr
	log.Println("hello, I am a test app. I will run until you signal me to stop!")
	fmt.Printf("the time is now: %s\n", start)
	go func() {
		<-time.After(time.Minute * 5)
		log.Fatalln("testapp ran for too long, killing process")
	}()
	if *keepalive == 1 {
		log.Printf("KeepAlive: APP_KEEPALIVE_PID_PATROL - we will NOT write PID: %d to file!", os.Getpid())
	} else if *keepalive == 2 {
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
	} else if *keepalive == 3 {
		log.Fatalln("KeepAlive: APP_KEEPALIVE_HTTP - NOT IMPLEMENTED")
	} else if *keepalive == 4 {
		log.Fatalln("KeepAlive: APP_KEEPALIVE_UDP - NOT IMPLEMENTED")
	} else {
		log.Fatalln("Unknown KeepAlive Method passed as a flag!")
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
				fmt.Println("Received SIGUSR1, ignoring!")
			case syscall.SIGUSR2:
				log.Println("Received SIGUSR2, ignoring!")
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
