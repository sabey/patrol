package main

import (
	"flag"
	"github.com/sabey/patrol"
	"log"
	"os"
	"os/signal"
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
	p, err := patrol.LoadPatrol(*config_path)
	if err != nil {
		log.Printf("./patrol/patrol.main(): failed to create Patrol: %s\n", err)
		os.Exit(255)
		return
	}
	// start patrol
	log.Println("./patrol/patrol.main(): Starting Patrol")
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
		sig := <-signals
		// handle signal
		// currently ALL of our options will result in a shutdown!!!
		// on shutdown we will notify our children of shutdown
		switch sig {
		case syscall.SIGHUP:
			// Hangup / ssh broken pipe
			log.Println("./patrol/patrol.main(): SIGHUP")
			done = true
			break
		case syscall.SIGINT:
			// terminate process
			// ctrl+c
			log.Println("./patrol/patrol.main(): SIGINT")
			done = true
			break
		case syscall.SIGQUIT:
			// ctrl+4 or ctrl+|
			log.Println("./patrol/patrol.main(): SIGQUIT")
			done = true
			break
		case syscall.SIGKILL:
			// kill -9
			// shutdown NOW
			log.Fatalf("./patrol/patrol.main(): SIGKILL")
			done = true
			break
		case syscall.SIGTERM:
			// killall service
			// gracefully shutdown NOW
			log.Println("./patrol/patrol.main(): SIGTERM")
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
			log.Printf("./patrol/patrol.main(): Unknown Signal Ignored: \"%v\"\n", sig)
		}
		if done {
			break
		}
	}
	log.Println("./patrol/patrol.main(): Stopping Patrol")
	p.Stop()
	// wait for patrol to stop
	log.Println("./patrol/patrol.main(): Waiting for Patrol to stop!")
	for {
		// we're going to add a saftey measure incase we fail to stop
		go func() {
			<-time.After(time.Minute * 3)
			log.Fatalln("./patrol/patrol.main(): Failed to Stop Patrol, Dying!")
		}()
		if !p.IsRunning() {
			// we're done!
			break
		}
	}
	log.Printf("./patrol/patrol.main(): Patrol ran for: %s\n", time.Now().Sub(start))
	log.Println("good bye!")
}
