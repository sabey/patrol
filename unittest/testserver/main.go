package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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

// this testserver should be run with `export PATROL_UNITTEST="I KNOW WHAT I AM DOING"`
func main() {
	start := time.Now()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	if !flag.Parsed() {
		flag.Parse()
	}
	// get wd
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("failed get working directory: %s\n", err)
		os.Exit(253)
		return
	}
	log.Printf("wd: \"%s\"\n", wd)
	config, err := patrol.LoadConfig(*config_path)
	if err != nil {
		log.Printf("failed to Load Patrol Config: %s\n", err)
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
	// WE HAVE TO SET OUR TIMESTAMP TO THE DEFAULT TIMESTAMP TYPE!!!
	config.Timestamp = time.RFC3339
	// we're going to override all of our triggers so that we can test for race conditions
	config.TriggerStart = func(
		p *patrol.Patrol,
	) error {
		// do nothing
		return nil
	}
	config.TriggerShutdown = func(
		p *patrol.Patrol,
	) {
		// do nothing
	}
	config.TriggerStarted = func(
		p *patrol.Patrol,
	) {
		// do nothing
	}
	config.TriggerTick = func(
		p *patrol.Patrol,
	) {
		// do nothing
	}
	config.TriggerStopped = func(
		p *patrol.Patrol,
	) {
		// do nothing
	}
	// override app triggers
	for _, a := range config.Apps {
		// overwrite stdout/err
		a.Stdout = os.Stdout
		a.Stderr = os.Stderr
		// prefix working directories!!!
		a.WorkingDirectory = filepath.Clean(wd + "/../" + a.WorkingDirectory)
		// add env
		a.Env = append(a.Env, fmt.Sprintf("%s=%s", patrol.PATROL_ENV_UNITTEST_KEY, patrol.PATROL_ENV_UNITTEST_VALUE))
		// triggers
		a.ExtraArgs = func(
			app *patrol.App,
		) []string {
			// do nothing
			return nil
		}
		a.ExtraEnv = func(
			app *patrol.App,
		) []string {
			// do nothing
			return nil
		}
		a.ExtraFiles = func(
			app *patrol.App,
		) []*os.File {
			// do nothing
			return nil
		}
		a.TriggerStart = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerStarted = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerStartedPinged = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerStartFailed = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerRunning = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerDisabled = func(
			app *patrol.App,
		) {
			// do nothing
		}
		a.TriggerClosed = func(
			app *patrol.App,
			history *patrol.History,
		) {
			// do nothing
		}
		a.TriggerPinged = func(
			app *patrol.App,
		) {
			// do nothing
		}
	}
	// override service triggers
	for _, s := range config.Services {
		s.TriggerStart = func(
			service *patrol.Service,
		) {
			// do nothing
		}
		s.TriggerStarted = func(
			service *patrol.Service,
		) {
			// do nothing
		}
		s.TriggerStartFailed = func(
			service *patrol.Service,
		) {
			// do nothing
		}
		s.TriggerRunning = func(
			service *patrol.Service,
		) {
			// do nothing
		}
		s.TriggerDisabled = func(
			service *patrol.Service,
		) {
			// do nothing
		}
		s.TriggerClosed = func(
			service *patrol.Service,
			history *patrol.History,
		) {
			// do nothing
		}
	}
	p, err = patrol.CreatePatrol(config)
	if err != nil {
		log.Printf("./patrol/unittest/testserver.main(): failed to Create Patrol: %s\n", err)
		os.Exit(255)
		return
	}
	// start patrol
	log.Println("./patrol/unittest/testserver.main(): Starting Patrol")
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
			log.Println("./patrol/unittest/testserver.main(): listener shutdown!")
			done = true
			break
		case sig := <-signals:
			switch sig {
			case syscall.SIGHUP:
				// Hangup / ssh broken pipe
				log.Println("./patrol/unittest/testserver.main(): SIGHUP")
				done = true
				break
			case syscall.SIGINT:
				// terminate process
				// ctrl+c
				log.Println("./patrol/unittest/testserver.main(): SIGINT")
				done = true
				break
			case syscall.SIGQUIT:
				// ctrl+4 or ctrl+|
				log.Println("./patrol/unittest/testserver.main(): SIGQUIT")
				done = true
				break
			case syscall.SIGKILL:
				// kill -9
				// shutdown NOW
				log.Println("./patrol/unittest/testserver.main(): SIGKILL")
				done = true
				break
			case syscall.SIGTERM:
				// killall service
				// gracefully shutdown NOW
				log.Println("./patrol/unittest/testserver.main(): SIGTERM")
				done = true
				break
			case syscall.SIGTSTP:
				// this will cause the program to go to the background if in a cli
				// ctrl+z
				log.Println("./patrol/unittest/testserver.main(): SIGTSTP")
				done = true
				break
			case syscall.SIGUSR1:
				// unreserved signal - handle however we want
				log.Println("./patrol/unittest/testserver.main(): SIGUSR1 - Ignored")
			case syscall.SIGUSR2:
				// unreserved signal - handle however we want
				log.Println("./patrol/unittest/testserver.main(): SIGUSR2 - Ignored")
			default:
				// unknown
				// do nothing
				log.Printf("./patrol/unittest/testserver.main(): Unknown Signal Ignored: \"%v\"\n", sig)
			}
		}
		if done {
			break
		}
	}
	log.Println("./patrol/unittest/testserver.main(): Stopping Patrol")
	p.Shutdown()
	// wait for patrol to stop
	log.Println("./patrol/unittest/testserver.main(): Waiting for Patrol to stop!")
	for {
		// we're going to add a saftey measure incase we fail to stop
		go func() {
			<-time.After(time.Minute * 3)
			log.Fatalln("./patrol/unittest/testserver.main(): Failed to Stop Patrol, Dying!")
		}()
		if !p.IsRunning() {
			// we're done!
			break
		}
	}
	log.Printf("./patrol/unittest/testserver.main(): Patrol ran for: %s\n", time.Now().Sub(start))
	log.Println("good bye!")
}
