package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sabey.co/patrol"
	"sync"
	"syscall"
	"time"
)

var (
	config_path = flag.String("config", "config.json", "path to patrol config file")
)
var (
	p           *patrol.Patrol
	shutdown_c  = make(chan struct{})
	shutdown_is bool
	shutdown_mu sync.Mutex
)

func shutdown() {
	shutdown_mu.Lock()
	defer shutdown_mu.Unlock()
	if shutdown_is {
		return
	}
	shutdown_is = true
	close(shutdown_c)
}

func main() {
	start := time.Now()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	if !flag.Parsed() {
		flag.Parse()
	}
	config, err := patrol.LoadConfig(*config_path)
	if err != nil {
		log.Printf("./patrol/patrol.main(): failed to Load Patrol Config: %s\n", err)
		os.Exit(254)
		return
	}
	// fix listeners
	// http
	if config.HTTP == nil {
		config.HTTP = &patrol.ConfigHTTP{
			Listen: httpListen(),
		}
	}
	if config.HTTP.Listen == "" {
		config.HTTP.Listen = httpListen()
	}
	if len(config.ListenHTTP) == 0 {
		config.ListenHTTP = []string{config.HTTP.Listen}
	}
	// udp
	if config.UDP == nil {
		config.UDP = &patrol.ConfigUDP{
			Listen: udpListen(),
		}
	}
	if config.UDP.Listen == "" {
		config.UDP.Listen = udpListen()
	}
	if len(config.ListenUDP) == 0 {
		config.ListenUDP = []string{config.UDP.Listen}
	}
	// modify timestamp
	config.Timestamp = time.RFC1123Z
	p, err = patrol.CreatePatrol(config)
	if err != nil {
		log.Printf("./patrol/patrol.main(): failed to Create Patrol: %s\n", err)
		os.Exit(255)
		return
	}
	// start patrol
	log.Println("./patrol/patrol.main(): Starting Patrol")
	p.Start()
	go HTTP()
	go UDP()
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
		// handle signals and shutdown
		select {
		case <-shutdown_c:
			log.Println("./unittest/testserver.main(): listener shutdown!")
			done = true
			break
		case sig := <-signals:
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
				log.Println("./patrol/patrol.main(): SIGKILL")
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
				log.Println("./patrol/patrol.main(): SIGTSTP")
				done = true
				break
			case syscall.SIGUSR1:
				// unreserved signal - handle however we want
				log.Println("./patrol/patrol.main(): SIGUSR1 - Ignored")
			case syscall.SIGUSR2:
				// unreserved signal - handle however we want
				log.Println("./patrol/patrol.main(): SIGUSR2 - Ignored")
			case syscall.SIGCHLD:
				// we want to ignore syscall.SIGCHLD
				// this is only used if we're using os.Exec and our child process exits
				// we're going to be SPAMMED with this when we're using Patrol
			default:
				// unknown
				// do nothing
				log.Printf("./patrol/patrol.main(): Unknown Signal Ignored: \"%v\"\n", sig)
			}
		}
		if done {
			break
		}
	}
	log.Println("./patrol/patrol.main(): Stopping Patrol")
	p.Shutdown()
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
