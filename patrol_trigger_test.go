package patrol

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolTriggerServices(t *testing.T) {
	log.Println("TestPatrolTriggerServices")

	patrol := &Patrol{}
	patrol.Apps = make(map[string]*PatrolApp)
	patrol.Services = make(map[string]*PatrolService)
	patrol.Services["ssh"] = &PatrolService{
		Name:       "ssh",
		Service:    "ssh",
		Management: SERVICE_MANAGEMENT_SERVICE,
	}
	// valid config!
	unittest.IsNil(t, patrol.validate())

	// trigger

	// add a service
	patrol.ServicesTrigger = make(map[string]*TriggerService)
	// empty Service
	patrol.ServicesTrigger[""] = &TriggerService{}
	// check that key exists
	_, exists := patrol.ServicesTrigger[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, patrol.validate(), ERR_SERVICESTRIGGER_KEY_EMPTY)

	// delete empty key
	delete(patrol.ServicesTrigger, "")
	_, exists = patrol.ServicesTrigger[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	patrol.ServicesTrigger["1234567890123456790123456789012345678901234567890123456789012345"] = &TriggerService{}
	unittest.Equals(t, patrol.validate(), ERR_SERVICESTRIGGER_KEY_INVALID)

	// delete invalid key
	delete(patrol.ServicesTrigger, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = patrol.ServicesTrigger["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	// nil
	patrol.ServicesTrigger["abc"] = nil
	unittest.Equals(t, patrol.validate(), ERR_SERVICESTRIGGER_TRIGGER_NIL)

	// valid
	patrol.ServicesTrigger["abc"] = &TriggerService{}
	unittest.IsNil(t, patrol.validate())
}
func TestPatrolTriggerApps(t *testing.T) {
	log.Println("TestPatrolTriggerApps")

	patrol := &Patrol{}
	patrol.Apps = make(map[string]*PatrolApp)
	patrol.Services = make(map[string]*PatrolService)

	patrol.Apps["http"] = &PatrolApp{
		KeepAlive:        APP_KEEPALIVE_PID_APP,
		Name:             "http",
		Binary:           "http",
		WorkingDirectory: "/directory",
		LogDirectory:     "logs",
		PIDPath:          "pid",
	}
	// valid config!
	unittest.IsNil(t, patrol.validate())

	// trigger

	// add a service
	patrol.AppsTrigger = make(map[string]*TriggerApp)
	// empty Service
	patrol.AppsTrigger[""] = &TriggerApp{}
	// check that key exists
	_, exists := patrol.AppsTrigger[""]
	unittest.Equals(t, exists, true)
	unittest.Equals(t, patrol.validate(), ERR_APPSTRIGGER_KEY_EMPTY)

	// delete empty key
	delete(patrol.AppsTrigger, "")
	_, exists = patrol.AppsTrigger[""]
	unittest.Equals(t, exists, false)

	// check for invalid key
	patrol.AppsTrigger["1234567890123456790123456789012345678901234567890123456789012345"] = &TriggerApp{}
	unittest.Equals(t, patrol.validate(), ERR_APPSTRIGGER_KEY_INVALID)

	// delete invalid key
	delete(patrol.AppsTrigger, "1234567890123456790123456789012345678901234567890123456789012345")
	_, exists = patrol.AppsTrigger["1234567890123456790123456789012345678901234567890123456789012345"]
	unittest.Equals(t, exists, false)

	// nil
	patrol.AppsTrigger["abc"] = nil
	unittest.Equals(t, patrol.validate(), ERR_APPSTRIGGER_TRIGGER_NIL)

	// valid
	patrol.AppsTrigger["abc"] = &TriggerApp{}
	unittest.IsNil(t, patrol.validate())
}
