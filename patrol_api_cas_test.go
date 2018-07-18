package patrol

import (
	"log"
	"os"
	"sabey.co/unittest"
	"testing"
	"time"
)

func TestAPICAS(t *testing.T) {
	log.Println("TestAPICAS")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	config := &Config{
		History:   5,
		Timestamp: time.RFC1123Z,
		Apps: map[string]*ConfigApp{
			"http": &ConfigApp{
				Name:             "testapp",
				KeepAlive:        APP_KEEPALIVE_HTTP,
				WorkingDirectory: wd + "/unittest/testapp",
				PIDPath:          "testapp.pid",
				LogDirectory:     "logs",
				Binary:           "testapp",
				// we're going to hijack our stderr and stdout for easy debugging
				Stderr: os.Stderr,
				Stdout: os.Stdout,
			},
		},
		Services: map[string]*ConfigService{
			"ssh": &ConfigService{
				Management: SERVICE_MANAGEMENT_SERVICE,
				Name:       "ssh",
				Service:    "ssh",
			},
		},
		ListenHTTP: []string{"127.0.0.1"},
	}
	// test http listeners
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	// ping
	request := &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
	}
	result := patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 0)
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	cas := result.CAS

	// ping again
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas+1)
	cas = result.CAS

	// ping again
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas+1)
	cas = result.CAS

	// NOOP
	// we have to increment cas since we know our previous CAS was modified
	cas++
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas)

	// NOOP again
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas)

	// NOOP again
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas)

	// CAS keyvalue
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
		CAS:   cas,
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.CAS, cas)

	// CAS keyvalue
	// cas was incremented
	cas++
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
		CAS:   cas,
		KeyValue: map[string]interface{}{
			"c": "d",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["a"], "b")
	unittest.Equals(t, result.CAS, cas)

	// CAS Ping
	cas++
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
		CAS:   cas,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.KeyValue["a"], "b")
	unittest.Equals(t, result.KeyValue["c"], "d")
	unittest.Equals(t, result.CAS, cas)

	// CAS Ping
	cas++
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
		CAS:   cas,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.KeyValue["a"], "b")
	unittest.Equals(t, result.KeyValue["c"], "d")
	unittest.Equals(t, result.CAS, cas)

	// NOOP
	// we have to increment cas since we know our previous CAS was modified
	cas++
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.CAS, cas)

	// NOOP again
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.CAS, cas)

	// NOOP again
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.CAS, cas)

	// cas invalid
	cas = result.CAS
	request = &API_Request{
		ID:    "http",
		Group: "app",
		CAS:   cas - 1,
		KeyValue: map[string]interface{}{
			"e": "f",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, true)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.CAS, cas)

	// NOOP
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.CASInvalid, false)
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.CAS, cas)
}
