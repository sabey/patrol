package main

import (
	"log"
	"os"
	"sabey.co/unittest"
	"testing"
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

	pid, err := app.getPID()
	unittest.IsNil(t, err)
	unittest.Equals(t, pid, 1254)

	app.PIDPath = "bad.pid"
	unittest.IsNil(t, app.validate())
	_, err = app.getPID()
	unittest.NotNil(t, err)
}
