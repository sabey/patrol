package patrol

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolServices(t *testing.T) {
	log.Println("TestPatrolServices")

	unittest.Equals(t, APP_NAME_MAXLENGTH, 255)

	config := &Config{}

	unittest.Equals(t, config.Validate(), ERR_PATROL_EMPTY)
	// Apps must be initialized, the creation of the config object will not do this for you
	config.Apps = make(map[string]*ConfigApp)
	unittest.Equals(t, config.Validate(), ERR_PATROL_EMPTY)

	// add a service
	config.Services = make(map[string]*ConfigService)
	// empty Service
	config.Services[""] = &ConfigService{}
	// check that key exists
	_, exists := config.Services[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, config.Validate(), ERR_SERVICES_KEY_EMPTY)

	// delete empty key
	delete(config.Services, "")
	_, exists = config.Services[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	config.Services["1234567890123456790123456789012345678901234567890123456789012345"] = &ConfigService{
		Management: SERVICE_MANAGEMENT_INITD,
	}
	unittest.Equals(t, config.Validate(), ERR_SERVICES_KEY_INVALID)

	// delete invalid key
	delete(config.Services, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = config.Services["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	service := &ConfigService{
		// empty object
	}
	config.Services["ssh"] = service
	// we're no longer going to get a config error, so we're good!
	unittest.Equals(t, config.Validate(), ERR_SERVICE_MANAGEMENT_INVALID)

	// create a valid service
	service.Service = "SSH"
	service.Name = "SSH Service"
	service.Management = SERVICE_MANAGEMENT_SERVICE

	// valid config!
	unittest.IsNil(t, config.Validate())

	// create patrol
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	// add duplicate case insensitive key
	// we can reuse our http object
	config.Services["SSH"] = service
	unittest.Equals(t, len(config.Services), 2)

	// Validate
	unittest.Equals(t, config.Validate(), ERR_SERVICE_LABEL_DUPLICATE)

	// delete service so we can Validate App
	config.Services = make(map[string]*ConfigService)
	unittest.Equals(t, config.Validate(), ERR_PATROL_EMPTY)
}
func TestPatrolApps(t *testing.T) {
	log.Println("TestPatrolApps")

	config := &Config{}

	unittest.Equals(t, config.Validate(), ERR_PATROL_EMPTY)
	// Apps must be initialized, the creation of the config object will not do this for you
	config.Apps = make(map[string]*ConfigApp)
	unittest.Equals(t, config.Validate(), ERR_PATROL_EMPTY)

	// add a service
	config.Services = make(map[string]*ConfigService)

	// empty App
	config.Apps[""] = &ConfigApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	// check that key exists
	_, exists := config.Apps[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, config.Validate(), ERR_APPS_KEY_EMPTY)

	// delete empty key
	delete(config.Apps, "")
	_, exists = config.Apps[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	config.Apps["1234567890123456790123456789012345678901234567890123456789012345"] = &ConfigApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	unittest.Equals(t, config.Validate(), ERR_APPS_KEY_INVALID)

	// delete invalid key
	delete(config.Apps, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = config.Apps["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	// valid object
	app := &ConfigApp{
		// empty object
	}
	config.Apps["http"] = app

	// no keep alive
	unittest.Equals(t, config.Validate(), ERR_APP_KEEPALIVE_INVALID)
	app.KeepAlive = APP_KEEPALIVE_PID_PATROL

	unittest.Equals(t, config.Validate(), ERR_APP_NAME_EMPTY)
	app.Name = "name"

	unittest.Equals(t, config.Validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)
	app.WorkingDirectory = "/directory"

	unittest.Equals(t, config.Validate(), ERR_APP_BINARY_EMPTY)
	app.Binary = "file"

	unittest.Equals(t, config.Validate(), ERR_APP_LOGDIRECTORY_EMPTY)
	app.LogDirectory = "log-directory"

	unittest.Equals(t, config.Validate(), ERR_APP_PIDPATH_EMPTY)
	app.PIDPath = "pid"

	// valid config!
	unittest.IsNil(t, config.Validate())

	// create patrol
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	// add duplicate case insensitive key
	// we can reuse our http object
	config.Apps["HTTP"] = app
	unittest.Equals(t, len(config.Apps), 2)

	// Validate
	unittest.Equals(t, config.Validate(), ERR_APP_LABEL_DUPLICATE)
}
