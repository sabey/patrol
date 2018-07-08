package patrol

import (
	"fmt"
	"log"
	"os"
	"sabey.co/unittest"
	"syscall"
	"testing"
	"time"
)

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
	unittest.Equals(t, app.history[0].Started.IsZero(), false)
	unittest.Equals(t, app.history[0].Stopped.IsZero(), false)
	unittest.Equals(t, app.history[0].Shutdown, false)
}
