package patrol

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolServices(t *testing.T) {
	log.Println("TestPatrolServices")

	unittest.Equals(t, APP_NAME_MAXLENGTH, 255)

	patrol := &Patrol{}

	unittest.Equals(t, patrol.validate(), ERR_PATROL_EMPTY)
	// Apps must be initialized, the creation of the patrol object will not do this for you
	patrol.Apps = make(map[string]*PatrolApp)
	unittest.Equals(t, patrol.validate(), ERR_PATROL_EMPTY)

	// add a service
	patrol.Services = make(map[string]*PatrolService)
	// empty Service
	patrol.Services[""] = &PatrolService{}
	// check that key exists
	_, exists := patrol.Services[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, patrol.validate(), ERR_SERVICES_KEY_EMPTY)

	// delete empty key
	delete(patrol.Services, "")
	_, exists = patrol.Services[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	patrol.Services["1234567890123456790123456789012345678901234567890123456789012345"] = &PatrolService{
		Management: SERVICE_MANAGEMENT_INITD,
	}
	unittest.Equals(t, patrol.validate(), ERR_SERVICES_KEY_INVALID)

	// delete invalid key
	delete(patrol.Services, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = patrol.Services["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	service := &PatrolService{
		// empty object
	}
	patrol.Services["ssh"] = service
	// we're no longer going to get a patrol error, so we're good!
	unittest.Equals(t, patrol.validate(), ERR_SERVICE_MANAGEMENT_INVALID)

	// create a valid service
	service.Service = "SSH"
	service.Name = "SSH Service"
	service.Management = SERVICE_MANAGEMENT_SERVICE

	// valid config!
	unittest.IsNil(t, patrol.validate())

	// add duplicate case insensitive key
	// we can reuse our http object
	patrol.Services["SSH"] = service
	unittest.Equals(t, len(patrol.Services), 2)

	// validate
	unittest.Equals(t, patrol.validate(), ERR_SERVICE_LABEL_DUPLICATE)

	// delete service so we can validate App
	patrol.Services = make(map[string]*PatrolService)
	unittest.Equals(t, patrol.validate(), ERR_PATROL_EMPTY)
}
func TestPatrolApps(t *testing.T) {
	log.Println("TestPatrolApps")

	patrol := &Patrol{}

	unittest.Equals(t, patrol.validate(), ERR_PATROL_EMPTY)
	// Apps must be initialized, the creation of the patrol object will not do this for you
	patrol.Apps = make(map[string]*PatrolApp)
	unittest.Equals(t, patrol.validate(), ERR_PATROL_EMPTY)

	// add a service
	patrol.Services = make(map[string]*PatrolService)

	// empty App
	patrol.Apps[""] = &PatrolApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	// check that key exists
	_, exists := patrol.Apps[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, patrol.validate(), ERR_APPS_KEY_EMPTY)

	// delete empty key
	delete(patrol.Apps, "")
	_, exists = patrol.Apps[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	patrol.Apps["1234567890123456790123456789012345678901234567890123456789012345"] = &PatrolApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	unittest.Equals(t, patrol.validate(), ERR_APPS_KEY_INVALID)

	// delete invalid key
	delete(patrol.Apps, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = patrol.Apps["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	// valid object
	app := &PatrolApp{
		// empty object
	}
	patrol.Apps["http"] = app

	// no keep alive
	unittest.Equals(t, patrol.validate(), ERR_APP_KEEPALIVE_INVALID)
	app.KeepAlive = APP_KEEPALIVE_PID_PATROL

	unittest.Equals(t, patrol.validate(), ERR_APP_NAME_EMPTY)
	app.Name = "name"

	unittest.Equals(t, patrol.validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)
	app.WorkingDirectory = "/directory"

	unittest.Equals(t, patrol.validate(), ERR_APP_BINARY_EMPTY)
	app.Binary = "file"

	unittest.Equals(t, patrol.validate(), ERR_APP_LOGDIRECTORY_EMPTY)
	app.LogDirectory = "log-directory"

	unittest.Equals(t, patrol.validate(), ERR_APP_PIDPATH_EMPTY)
	app.PIDPath = "pid"

	// valid config!
	unittest.IsNil(t, patrol.validate())

	// add duplicate case insensitive key
	// we can reuse our http object
	patrol.Apps["HTTP"] = app
	unittest.Equals(t, len(patrol.Apps), 2)

	// validate
	unittest.Equals(t, patrol.validate(), ERR_APP_LABEL_DUPLICATE)

}
