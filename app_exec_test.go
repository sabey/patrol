package patrol

import (
	"encoding/json"
	"fmt"
	"log"
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
	fmt.Printf("signalling app PID: %d\n", pid)
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
	fmt.Printf("signalling app PID: %d\n", pid)
	app.signalStop()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")
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
}
func TestAppExecPatrolDisable(t *testing.T) {
	log.Println("TestAppExecPatrolDisable")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
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
	fmt.Printf("signalling app PID: %d\n", pid)
	app.signalStop()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")
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
}
func TestAppExecPatrolLogDirectory(t *testing.T) {
	log.Println("TestAppExecPatrolLogDirectory")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &App{
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
	fmt.Printf("signalling app PID: %d\n", pid)
	app.signalStop()

	// wait for our process to be killed
	fmt.Println("waiting for app to be killed")
	<-time.After(time.Second * 2)
	fmt.Println("app closed")
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
	fmt.Printf("signalling app PID: %d\n", app.GetPID())
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
	// check that our process is dead
	unittest.NotNil(t, app.isAppRunning())
	// check our history

	unittest.Equals(t, len(app.history), 1)
	// PID should exist, but may not, it's not supported with APP_PID
	// unittest.Equals(t, app.history[0].PID != 0, true)
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
}
