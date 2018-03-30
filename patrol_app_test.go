package main

import (
	"fmt"
	"log"
	"os"
	"sabey.co/unittest"
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
	unittest.Equals(t, app.validate(), ERR_APP_APPPATH_EMPTY)

	// setting working directory to cwd
	app.WorkingDirectory = wd + "/unittest"

	app.AppPath = "file/.."
	unittest.Equals(t, app.validate(), ERR_APP_APPPATH_UNCLEAN)

	app.AppPath = "file"
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
func TestPatrolAppTestApp(t *testing.T) {
	log.Println("TestPatrolAppTestApp")
	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)
	app := &PatrolApp{
		Name:             "testapp",
		KeepAlive:        APP_KEEPALIVE_PID_APP,
		WorkingDirectory: wd + "/unittest/testapp",
		PIDPath:          "testapp.pid",
		LogDirectory:     "logs",
		AppPath:          "testapp",
	}
	unittest.NotNil(t, app.isAppRunning())
	unittest.IsNil(t, app.startApp())
	<-time.After(time.Second * 2)
	unittest.IsNil(t, app.isAppRunning())
}
