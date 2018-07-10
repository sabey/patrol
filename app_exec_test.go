package patrol

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sabey.co/unittest"
	"syscall"
	"testing"
	"time"
)

func TestAppExecPatrolPID(t *testing.T) {
	log.Println("TestAppExecPatrolPID")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
		id: "testapp",
		// this must be set or we will get an an error when saving history
		patrol: &Patrol{
			config: &Config{
				History:   5,
				Timestamp: time.RFC1123Z,
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
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	unittest.IsNil(t, app.isAppRunning())
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
	app.mu.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	unittest.Equals(t, app.history[0].PID, pid)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.IsNil(t, app.history[0].LastSeen)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
	unittest.Equals(t, app.history[0].ExitCode, 0)
	app.mu.Unlock()

	// test timestamp marshal
	bs1, _ := json.MarshalIndent(app.history, "", "\t")

	// we can't use []interface{}{map[string]struct{}{} because when we scan over our map it isn't deterministic
	result := []interface{}{
		struct {
			PID     uint32 `json:"pid,omitempty"`
			Started string `json:"started,omitempty"`
			Stopped string `json:"stopped,omitempty"`
		}{
			PID:     pid,
			Started: app.history[0].Started.Format(app.patrol.config.Timestamp),
			Stopped: app.history[0].Stopped.Format(app.patrol.config.Timestamp),
		},
	}
	bs2, _ := json.MarshalIndent(result, "", "\t")

	unittest.Equals(t, string(bs1), string(bs2))
}
func TestAppExecPatrolShutdown(t *testing.T) {
	log.Println("TestAppExecPatrolShutdown")

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
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	unittest.IsNil(t, app.isAppRunning())
	fmt.Println("waited for app")

	// mark as shutdown
	app.patrol.shutdown = true

	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.mu.Lock()
	app.signalStop()
	app.mu.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.mu.Lock()
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
	app.mu.Unlock()
}
func TestAppExecPatrolDisable(t *testing.T) {
	log.Println("TestAppExecPatrolDisable")

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
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	unittest.IsNil(t, app.isAppRunning())
	fmt.Println("waited for app")

	// set disabled
	app.Disable()
	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.mu.Lock()
	app.signalStop()
	app.mu.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.mu.Lock()
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
	app.mu.Unlock()
}
func TestAppExecPatrolLogDirectory(t *testing.T) {
	log.Println("TestAppExecPatrolLogDirectory")

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
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	unittest.IsNil(t, app.isAppRunning())
	fmt.Println("waited for app")

	// set disabled
	app.Disable()
	// save pid
	pid := app.GetPID()
	fmt.Printf("signaling app PID: %d\n", pid)

	app.mu.Lock()
	app.signalStop()
	app.mu.Unlock()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")

	// as soon as our app is closed we have to use a mutex since our closing function runs in a goroutine
	app.mu.Lock()
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
	app.mu.Unlock()

	// verify our logs exist

	// stdout
	f1, err := os.Open(fmt.Sprintf("%s/%d.stdout.log", app.logDir(), app.history[0].Started.UnixNano()))
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
	f2, err := os.Open(fmt.Sprintf("%s/%d.stderr.log", app.logDir(), app.history[0].Started.UnixNano()))
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
	}

	// this will fail if testapp is somehow running
	// testapp has a self destruct function, it should be about 30 seconds
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 3)
	unittest.IsNil(t, app.isAppRunning())
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
	app.mu.Lock()
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history
	unittest.Equals(t, len(app.history), 1)
	// PID should exist, but may not, it's not supported with APP_PID
	// unittest.Equals(t, app.history[0].PID != 0, true)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
	app.mu.Unlock()
}
func TestAppExecHTTP(t *testing.T) {
	log.Println("TestAppExecHTTP")

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
		Timestamp:   time.RFC1123Z,
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
		mux.HandleFunc("/ping/", patrol.ServeHTTPPing)
		// serve will block
		http.Serve(l, mux)
	}()
	// wait a second for http server to be built
	<-time.After(time.Second * 1)

	// start testapp
	// not running
	unittest.NotNil(t, patrol.apps["http"].isAppRunning())
	unittest.IsNil(t, patrol.apps["http"].startApp())
	// from here on out we have to use a mutex incase ping races
	// app should be running now
	// our comparator should be based off of started until we receive a ping
	// once we receive our first ping we will compare on ping
	patrol.apps["http"].mu.Lock()
	unittest.IsNil(t, patrol.apps["http"].isAppRunning())
	patrol.apps["http"].mu.Unlock()
	// we have to wait a second or two for Start() to run AND THEN have testapp write our PID to file
	// if we do not wait testapp could run and not yet write PID
	fmt.Println("waiting for app")
	<-time.After(time.Second * 5)
	fmt.Println("waited for app")
	// verify we're receiving pings
	patrol.apps["http"].mu.Lock()
	unittest.IsNil(t, patrol.apps["http"].isAppRunning())
	// verify we've received a PID back
	pid := patrol.apps["http"].pid
	unittest.Equals(t, pid > 0, true)
	patrol.apps["http"].mu.Unlock()

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
	patrol.apps["http"].mu.Lock()
	unittest.NotNil(t, patrol.apps["http"].isAppRunning())
	// check our history
	unittest.Equals(t, len(patrol.apps["http"].history), 1)
	unittest.Equals(t, patrol.apps["http"].history[0].PID, pid)
	unittest.Equals(t, patrol.apps["http"].history[0].Started.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].LastSeen.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].Stopped.IsZero(), false)
	unittest.Equals(t, patrol.apps["http"].history[0].Shutdown, false)
	unittest.Equals(t, patrol.apps["http"].history[0].ExitCode, 0)
	patrol.apps["http"].mu.Unlock()

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
