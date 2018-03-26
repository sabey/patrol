package main

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrol(t *testing.T) {
	log.Println("TestPatrol")

	unittest.Equals(t, APP_NAME_MAXLENGTH, 255)

	patrol := &Patrol{}

	unittest.Equals(t, patrol.validate(), ERR_APPS_EMPTY)
	// Apps must be initialized, the creation of the patrol object will not do this for you
	patrol.Apps = make(map[string]*PatrolApp)
	unittest.Equals(t, patrol.validate(), ERR_APPS_EMPTY)

	patrol.Apps[""] = &PatrolApp{}
	// check that key exists
	_, exists := patrol.Apps[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, patrol.validate(), ERR_APP_KEY_EMPTY)

	// delete empty key
	delete(patrol.Apps, "")
	_, exists = patrol.Apps[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	patrol.Apps["123456789012345679012345678912345"] = &PatrolApp{}
	unittest.Equals(t, patrol.validate(), ERR_APP_KEY_INVALID)

	// delete invalid key
	delete(patrol.Apps, "123456789012345679012345678912345")
	_, exists = patrol.Apps["123456789012345679012345678912345"]
	unittest.Equals(t, exists, false)

	app := &PatrolApp{
		// empty object
	}
	patrol.Apps["http"] = app

	unittest.Equals(t, patrol.validate(), ERR_APP_NAME_EMPTY)
	app.Name = "name"

	unittest.Equals(t, patrol.validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)
	app.WorkingDirectory = "directory"

	unittest.Equals(t, patrol.validate(), ERR_APP_APPPATH_EMPTY)
	app.AppPath = "file"

	unittest.Equals(t, patrol.validate(), ERR_APP_LOG_DIRECTORY_EMPTY)
	app.LogDirectory = "log-directory"

	unittest.Equals(t, patrol.validate(), ERR_APP_PIDPATH_EMPTY)
	app.PIDPath = "pid"

	unittest.IsNil(t, app.validate())
}
func TestPatrolApp(t *testing.T) {
	log.Println("TestPatrolApp")

	app := &PatrolApp{}
	unittest.Equals(t, app.validate(), ERR_APP_NAME_EMPTY)

	app.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, app.validate(), ERR_APP_NAME_MAXLENGTH)

	app.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, app.validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)

	app.WorkingDirectory = "directory/."
	unittest.Equals(t, app.validate(), ERR_APP_WORKINGDIRECTORY_UNCLEAN)

	app.WorkingDirectory = "directory"
	unittest.Equals(t, app.validate(), ERR_APP_APPPATH_EMPTY)

	app.AppPath = "file/.."
	unittest.Equals(t, app.validate(), ERR_APP_APPPATH_UNCLEAN)

	app.AppPath = "file"
	unittest.Equals(t, app.validate(), ERR_APP_LOG_DIRECTORY_EMPTY)

	app.LogDirectory = "log-directory/."
	unittest.Equals(t, app.validate(), ERR_APP_LOG_DIRECTORY_UNCLEAN)

	app.LogDirectory = "log-directory"
	unittest.Equals(t, app.validate(), ERR_APP_PIDPATH_EMPTY)

	app.PIDPath = "pid/.."
	unittest.Equals(t, app.validate(), ERR_APP_PIDPATH_UNCLEAN)

	app.PIDPath = "pid"

	unittest.IsNil(t, app.validate())
}