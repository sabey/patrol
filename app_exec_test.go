package patrol

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sabey.co/patrol/cas"
	"sabey.co/unittest"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestAppExecPatrolPID(t *testing.T) {
	log.Println("TestAppExecPatrolPID")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecPatrolPID")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History:   5,
				Timestamp: time.RFC3339,
			},
		},
		config: &ConfigApp{
			Name:             "testapp",
			KeepAlive:        APP_KEEPALIVE_PID_PATROL,
			WorkingDirectory: wd + "/unittest/testapp",
			PIDPath:          "testapp.pid",
			LogDirectory:     "logs",
			Binary:           "testapp",
			// we're going to hijack our stderr and stdout for easy debugging
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		},
		o: cas.CreateApp(false),
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	app.o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	app.o.Lock()
	unittest.IsNil(t, app.isAppRunning())
	app.o.Unlock()
	fmt.Println("waited for app")

	// we have to signal our app to stop, just incase we were to run this unittest again
	unittest.Equals(t, app.GetPID() > 0, true)
	process, err := os.FindProcess(int(app.GetPID()))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)
	// do NOT use process.Kill() - we're not able to catch the signal
	// this seems to be an alias for process.Signal(syscall.SIGKILL) ???
	// either way, our testapp never receives this signal, so it is useful for unittesting
	// we're just going to use either SIGINT or SIGHUP
	unittest.IsNil(t, process.Signal(syscall.SIGHUP))
	// alternatively we can write
	//unittest.IsNil(t, syscall.Kill(int(app.GetPID()), syscall.SIGINT))

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	// check that our process is dead
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	unittest.Equals(t, app.history[0].PID, pid)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].LastSeen.IsZero(), false) // last seen must always exist if we're running
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
	unittest.Equals(t, app.history[0].ExitCode, 0)
	app.o.Unlock()

	// test timestamp marshal
	bs1, _ := json.MarshalIndent(app.history, "", "\t")

	// we can't use []interface{}{map[string]struct{}{} because when we scan over our map it isn't deterministic
	result := []interface{}{
		struct {
			PID      uint32 `json:"pid,omitempty"`
			Started  string `json:"started,omitempty"`
			Lastseen string `json:"lastseen,omitempty"`
			Stopped  string `json:"stopped,omitempty"`
		}{
			PID:      pid,
			Started:  app.history[0].Started.Format(app.patrol.config.Timestamp),
			Lastseen: app.history[0].LastSeen.Format(app.patrol.config.Timestamp),
			Stopped:  app.history[0].Stopped.Format(app.patrol.config.Timestamp),
		},
	}
	bs2, _ := json.MarshalIndent(result, "", "\t")
	unittest.Equals(t, string(bs1), string(bs2))
}
func TestAppExecPatrolShutdown(t *testing.T) {
	log.Println("TestAppExecPatrolShutdown")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecPatrolShutdown")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History: 5,
			},
		},
		config: &ConfigApp{
			Name:             "testapp",
			KeepAlive:        APP_KEEPALIVE_PID_PATROL,
			WorkingDirectory: wd + "/unittest/testapp",
			PIDPath:          "testapp.pid",
			LogDirectory:     "logs",
			Binary:           "testapp",
			// we're going to hijack our stderr and stdout for easy debugging
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		},
		o: cas.CreateApp(false),
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	app.o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	app.o.Lock()
	unittest.IsNil(t, app.isAppRunning())
	app.o.Unlock()
	fmt.Println("waited for app")

	// mark as shutdown
	app.patrol.shutdown = true

	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.o.Lock()
	app.signalStop()
	app.o.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.o.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	unittest.Equals(t, app.history[0].PID, pid)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Disabled, false)
	unittest.Equals(t, app.history[0].Shutdown, true)
	// check that we were notified - SIGUSR1
	unittest.Equals(t, app.history[0].ExitCode, 10)
	app.o.Unlock()
}
func TestAppExecPatrolDisable(t *testing.T) {
	log.Println("TestAppExecPatrolDisable")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecPatrolDisable")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History: 5,
			},
		},
		config: &ConfigApp{
			Name:             "testapp",
			KeepAlive:        APP_KEEPALIVE_PID_PATROL,
			WorkingDirectory: wd + "/unittest/testapp",
			PIDPath:          "testapp.pid",
			LogDirectory:     "logs",
			Binary:           "testapp",
			// we're going to hijack our stderr and stdout for easy debugging
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		},
		o: cas.CreateApp(false),
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	app.o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	app.o.Lock()
	unittest.IsNil(t, app.isAppRunning())
	app.o.Unlock()
	fmt.Println("waited for app")

	// set disabled
	app.Disable()
	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.o.Lock()
	app.signalStop()
	app.o.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.o.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	unittest.Equals(t, app.history[0].PID, pid)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Disabled, true)
	unittest.Equals(t, app.history[0].Shutdown, false)
	// check that we were notified - SIGUSR2
	unittest.Equals(t, app.history[0].ExitCode, 12)
	app.o.Unlock()
}
func TestAppExecPatrolLogDirectory(t *testing.T) {
	log.Println("TestAppExecPatrolLogDirectory")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecPatrolLogDirectory")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History: 5,
			},
		},
		config: &ConfigApp{
			Name:             "testapp",
			KeepAlive:        APP_KEEPALIVE_PID_PATROL,
			WorkingDirectory: wd + "/unittest/testapp",
			PIDPath:          "testapp.pid",
			LogDirectory:     "logs",
			Binary:           "testapp",
			// we're NOTgoing to hijack our stderr and stdout!!!!
		},
		o: cas.CreateApp(false),
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	unittest.Equals(t, app.o.GetStartedLog().IsZero(), true)
	unittest.IsNil(t, app.startApp())
	unittest.Equals(t, app.o.GetStartedLog().IsZero(), false)
	app.o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	app.o.Lock()
	unittest.IsNil(t, app.isAppRunning())
	app.o.Unlock()
	fmt.Println("waited for app")

	// set disabled
	app.Disable()
	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.o.Lock()
	app.signalStop()
	app.o.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.o.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	unittest.Equals(t, app.o.GetStartedLog().IsZero(), false)
	// check our history
	unittest.Equals(t, len(app.history), 1)
	unittest.Equals(t, app.history[0].PID, pid)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Disabled, true)
	unittest.Equals(t, app.history[0].Shutdown, false)
	// check that we were notified - SIGUSR2
	unittest.Equals(t, app.history[0].ExitCode, 12)
	app.o.Unlock()

	// verify our logs exist

	app.o.Lock()
	s := fmt.Sprintf("%s/%d.stdout.log", app.logDir(), app.o.GetStartedLog().UnixNano())
	app.o.Unlock()
	// stdout
	f1, err := os.Open(s)
	unittest.IsNil(t, err)
	unittest.NotNil(t, f1)
	// read 2 bytes
	bs := make([]byte, 2)
	n, err := f1.Read(bs)
	unittest.IsNil(t, err)
	unittest.Equals(t, n, 2)
	unittest.Equals(t, len(bs), 2)
	// close
	unittest.IsNil(t, f1.Close())
	// stderr
	app.o.Lock()
	s = fmt.Sprintf("%s/%d.stderr.log", app.logDir(), app.o.GetStartedLog().UnixNano())
	app.o.Unlock()
	f2, err := os.Open(s)
	unittest.IsNil(t, err)
	unittest.NotNil(t, f2)
	// read 2 bytes
	bs = make([]byte, 2)
	n, err = f2.Read(bs)
	unittest.IsNil(t, err)
	unittest.Equals(t, n, 2)
	unittest.Equals(t, len(bs), 2)
	// close
	unittest.IsNil(t, f2.Close())
}
func TestAppExecAppPID(t *testing.T) {
	log.Println("TestAppExecAppPID")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecAppPID")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History: 5,
			},
		},
		config: &ConfigApp{
			Name:             "testapp",
			KeepAlive:        APP_KEEPALIVE_PID_APP,
			WorkingDirectory: wd + "/unittest/testapp",
			PIDPath:          "testapp.pid",
			LogDirectory:     "logs",
			Binary:           "testapp",
			// we're going to hijack our stderr and stdout for easy debugging
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		},
		o: cas.CreateApp(false),
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	app.o.Lock()
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 0)
	// start app
	unittest.IsNil(t, app.startApp())
	app.o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	app.o.Lock()
	unittest.IsNil(t, app.isAppRunning())
	app.o.Unlock()
	fmt.Println("waited for app")

	// we have to signal our app to stop, just incase we were to run this unittest again
	unittest.Equals(t, app.GetPID() > 0, true)
	process, err := os.FindProcess(int(app.GetPID()))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	fmt.Printf("signaling app PID: %d\n", app.GetPID())
	// do NOT use process.Kill() - we're not able to catch the signal
	// this seems to be an alias for process.Signal(syscall.SIGKILL) ???
	// either way, our testapp never receives this signal, so it is useful for unittesting
	// we're just going to use either SIGINT or SIGHUP
	unittest.IsNil(t, process.Signal(syscall.SIGHUP))
	// alternatively we can write
	//unittest.IsNil(t, syscall.Kill(int(app.GetPID()), syscall.SIGINT))

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.o.Lock()
	// check our history
	unittest.Equals(t, len(app.history), 1)
	// PID should exist, but may not, it's not supported with APP_PID
	// unittest.Equals(t, app.history[0].PID != 0, true)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
	app.o.Unlock()

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.o.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	// PID should exist, but may not, it's not supported with APP_PID
	// unittest.Equals(t, app.history[0].PID != 0, true)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
	app.o.Unlock()
}
func TestAppExecHTTP(t *testing.T) {
	log.Println("TestAppExecHTTP")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecHTTP")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	// we need to create a net listener on an unavailable port so we can pass it in our config
	l, err := net.Listen("tcp", ":0")
	unittest.IsNil(t, err)
	defer l.Close()
	// get address
	log.Printf("listening on: \"%s\"\n", l.Addr().String())

	config := &Config{
		History:     5,
		Timestamp:   time.RFC3339,
		PingTimeout: APP_PING_TIMEOUT_MIN, // we're going to overwrite this internally
		Apps: map[string]*ConfigApp{
			"http": &ConfigApp{
				Name:             "testapp",
				KeepAlive:        APP_KEEPALIVE_HTTP,
				WorkingDirectory: wd + "/unittest/testapp",
				PIDPath:          "testapp.pid",
				LogDirectory:     "logs",
				Binary:           "testapp",
				// we're going to hijack our stderr and stdout for easy debugging
				Stderr: os.Stderr,
				Stdout: os.Stdout,
			},
		},
		ListenHTTP: []string{
			l.Addr().String(),
		},
	}
	// test http listeners
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	// we're going to overwrite our ping timeout internally
	patrol.config.PingTimeout = 2

	// create http server
	go func() {
		// create mux
		mux := http.NewServeMux()
		mux.HandleFunc("/api/", patrol.ServeHTTPAPI)
		// serve will block
		http.Serve(l, mux)
	}()
	// wait a second for http server to be built
	<-time.After(time.Second * 1)

	// start testapp
	// not running
	patrol.apps["http"].o.Lock()
	unittest.NotNil(t, patrol.apps["http"].isAppRunning())
	unittest.IsNil(t, patrol.apps["http"].startApp())
	patrol.apps["http"].o.Unlock()
	// from here on out we have to use a mutex incase ping races
	// app should be running now
	// our comparator should be based off of started until we receive a ping
	// once we receive our first ping we will compare on ping
	patrol.apps["http"].o.Lock()
	unittest.IsNil(t, patrol.apps["http"].isAppRunning())
	patrol.apps["http"].o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 5)
	fmt.Println("waited for app")
	// verify we're receiving pings
	patrol.apps["http"].o.Lock()
	unittest.IsNil(t, patrol.apps["http"].isAppRunning())
	patrol.apps["http"].o.Unlock()
	// verify we've received a PID back
	pid := patrol.apps["http"].GetPID()
	unittest.Equals(t, pid > 0, true)

	// signal back we want pinging to stop
	fmt.Printf("signaling app PID: %d\n", pid)
	process, err := os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.SIGINT))

	// ping will timeout after 2 seconds
	// wait for keepalive to expire
	fmt.Println("waiting for app to timeout")
	<-time.After(time.Second * 5)
	fmt.Println("app timed out")

	// check that our process is considered to be not running
	patrol.apps["http"].o.Lock()
	unittest.NotNil(t, patrol.apps["http"].isAppRunning())
	// check our history
	unittest.Equals(t, len(patrol.apps["http"].history), 1)
	unittest.Equals(t, patrol.apps["http"].history[0].PID, pid)
	unittest.Equals(t, patrol.apps["http"].history[0].Started.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].LastSeen.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].Stopped.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].Shutdown, false)
	unittest.Equals(t, patrol.apps["http"].history[0].ExitCode, 0)
	patrol.apps["http"].o.Unlock()

	fmt.Println("verifying app is still alive")
	// our process SHOULD still be alive
	// we're going to use kill -0 to double check
	fmt.Printf("signaling kill -0 PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	fmt.Printf("signaling app to exit PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.SIGKILL))

	fmt.Println("waiting to exit")
	<-time.After(time.Second * 3)

	fmt.Println("verifying app is dead")
	// our process SHOULD still dead
	// we're going to use kill -0 to double check
	fmt.Printf("signaling kill -0 PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.NotNil(t, process.Signal(syscall.Signal(0)))
}
func TestAppExecUDP(t *testing.T) {
	log.Println("TestAppExecUDP")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestAppExecUDP")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	// we need to create a net listener on an unavailable port so we can pass it in our config
	c, err := net.ListenPacket("udp", ":0")
	unittest.IsNil(t, err)
	defer c.Close()
	conn, ok := c.(*net.UDPConn)
	unittest.Equals(t, ok, true)

	// get address
	log.Printf("listening on: \"%s\"\n", c.LocalAddr().String())

	config := &Config{
		History:     5,
		Timestamp:   time.RFC3339,
		PingTimeout: APP_PING_TIMEOUT_MIN, // we're going to overwrite this internally
		Apps: map[string]*ConfigApp{
			"udp": &ConfigApp{
				Name:             "testapp",
				KeepAlive:        APP_KEEPALIVE_UDP,
				WorkingDirectory: wd + "/unittest/testapp",
				PIDPath:          "testapp.pid",
				LogDirectory:     "logs",
				Binary:           "testapp",
				// we're going to hijack our stderr and stdout for easy debugging
				Stderr: os.Stderr,
				Stdout: os.Stdout,
			},
		},
		ListenUDP: []string{
			c.LocalAddr().String(),
		},
	}
	// test udp listeners
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	// we're going to overwrite our ping timeout internally
	patrol.config.PingTimeout = 2

	// create udp server
	udp_done := false
	var udp_mu sync.Mutex
	go func() {
		for {
			if err := patrol.HandleUDPConnection(conn); err != nil {
				// we failed to write
				udp_mu.Lock()
				if udp_done {
					// we're done!!!
					// ignore this error
					udp_mu.Unlock()
					return
				}
				udp_mu.Unlock()
				log.Fatalf("udp handle error: \"%s\"\n", err)
				return
			}
		}
	}()
	// wait a second for udp server to be built
	<-time.After(time.Second * 1)

	// start testapp
	// not running
	patrol.apps["udp"].o.Lock()
	unittest.NotNil(t, patrol.apps["udp"].isAppRunning())
	unittest.IsNil(t, patrol.apps["udp"].startApp())
	patrol.apps["udp"].o.Unlock()
	// from here on out we have to use a mutex incase ping races
	// app should be running now
	// our comparator should be based off of started until we receive a ping
	// once we receive our first ping we will compare on ping
	patrol.apps["udp"].o.Lock()
	unittest.IsNil(t, patrol.apps["udp"].isAppRunning())
	patrol.apps["udp"].o.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 5)
	fmt.Println("waited for app")
	// verify we're receiving pings
	patrol.apps["udp"].o.Lock()
	unittest.IsNil(t, patrol.apps["udp"].isAppRunning())
	patrol.apps["udp"].o.Unlock()
	// verify we've received a PID back
	pid := patrol.apps["udp"].GetPID()
	unittest.Equals(t, pid > 0, true)

	// signal back we want pinging to stop
	fmt.Printf("signaling app PID: %d\n", pid)
	process, err := os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.SIGINT))

	// ping will timeout after 2 seconds
	// wait for keepalive to expire
	fmt.Println("waiting for app to timeout")
	<-time.After(time.Second * 5)
	fmt.Println("app timed out")

	// check that our process is considered to be not running
	patrol.apps["udp"].o.Lock()
	unittest.NotNil(t, patrol.apps["udp"].isAppRunning())
	// check our history
	unittest.Equals(t, len(patrol.apps["udp"].history), 1)
	unittest.Equals(t, patrol.apps["udp"].history[0].PID, pid)
	unittest.Equals(t, patrol.apps["udp"].history[0].Started.IsZero(), false)
	unittest.Equals(t, patrol.apps["udp"].history[0].LastSeen.IsZero(), false)
	unittest.Equals(t, patrol.apps["udp"].history[0].Stopped.IsZero(), false)
	unittest.Equals(t, patrol.apps["udp"].history[0].Shutdown, false)
	unittest.Equals(t, patrol.apps["udp"].history[0].ExitCode, 0)
	patrol.apps["udp"].o.Unlock()

	fmt.Println("verifying app is still alive")
	// our process SHOULD still be alive
	// we're going to use kill -0 to double check
	fmt.Printf("signaling kill -0 PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	udp_mu.Lock()
	udp_done = true
	udp_mu.Unlock()

	fmt.Printf("signaling app to exit PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.SIGKILL))

	fmt.Println("waiting to exit")
	<-time.After(time.Second * 3)

	fmt.Println("verifying app is dead")
	// our process SHOULD still dead
	// we're going to use kill -0 to double check
	fmt.Printf("signaling kill -0 PID: %d\n", pid)
	process, err = os.FindProcess(int(pid))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.NotNil(t, process.Signal(syscall.Signal(0)))
}
