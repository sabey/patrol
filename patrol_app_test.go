package main

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolApp(t *testing.T) {
	log.Println("TestPatrolApp")

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

	app.PIDPath = "pid"
	unittest.IsNil(t, app.validate())
}
