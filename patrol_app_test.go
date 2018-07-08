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

func TestPatrolApp(t *testing.T) {
	log.Println("TestPatrolApp")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &PatrolApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	unittest.Equals(t, app.validate(), ERR_APP_NAME_EMPTY)

	app.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, app.validate(), ERR_APP_NAME_MAXLENGTH)

	app.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, app.validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)

	app.WorkingDirectory = "directory"
	unittest.Equals(t, app.validate(), ERR_APP_WORKINGDIRECTORY_RELATIVE)

	app.WorkingDirectory = "/directory/."
	unittest.Equals(t, app.validate(), ERR_APP_WORKINGDIRECTORY_UNCLEAN)

	app.WorkingDirectory = "/directory"
	unittest.Equals(t, app.validate(), ERR_APP_BINARY_EMPTY)

	// setting working directory to cwd
	app.WorkingDirectory = wd + "/unittest"

	app.Binary = "file/.."
	unittest.Equals(t, app.validate(), ERR_APP_BINARY_UNCLEAN)

	app.Binary = "file"
	unittest.Equals(t, app.validate(), ERR_APP_LOGDIRECTORY_EMPTY)

	app.LogDirectory = "log-directory/."
	unittest.Equals(t, app.validate(), ERR_APP_LOGDIRECTORY_UNCLEAN)

	app.LogDirectory = "log-directory"
	unittest.Equals(t, app.validate(), ERR_APP_PIDPATH_EMPTY)

	app.PIDPath = "pid/.."
	unittest.Equals(t, app.validate(), ERR_APP_PIDPATH_UNCLEAN)

	// changing pid to app.pid
	app.PIDPath = "app.pid"

	unittest.IsNil(t, app.validate())

	fmt.Println("app.getPID")

	pid, err := app.getPID()
	unittest.IsNil(t, err)
	unittest.Equals(t, pid, 1254)

	app.PIDPath = "bad.pid"
	unittest.IsNil(t, app.validate())
	_, err = app.getPID()
	unittest.NotNil(t, err)
}
func TestPatrolAppTestAppPIDAPP(t *testing.T) {
	log.Println("TestPatrolAppTestAppPIDAPP")

	// set our internal unittesting variable
	unittesting = true
	defer func() {
		unittesting = false
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	app := &PatrolApp{
		Name:             "testapp",
		KeepAlive:        APP_KEEPALIVE_PID_APP,
		WorkingDirectory: wd + "/unittest/testapp",
		PIDPath:          "testapp.pid",
		LogDirectory:     "logs",
		Binary:           "testapp",
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
}
