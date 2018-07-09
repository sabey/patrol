package patrol

import (
	"log"
	"os"
	"sabey.co/unittest"
	"testing"
	"time"
)

func TestAPI(t *testing.T) {
	log.Println("TestAPI")

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
	}
	patrol, err := CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	result := patrol.API(nil)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Request NIL")
	result = patrol.Ping(nil)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Request NIL")

	request := &API_Request{}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")

	request = &API_Request{
		Group: "?",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")

	request = &API_Request{
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")

	request = &API_Request{
		Group: "apps",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")

	request = &API_Request{
		Group: "service",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Service")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Services don't support Ping")

	request = &API_Request{
		Group: "services",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Service")
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Services don't support Ping")

	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Toggle: API_TOGGLE_DISABLE,
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	t1 := result.LastSeen.Time

	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Toggle: API_TOGGLE_ENABLE,
		PID:    0,
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	t2 := result.LastSeen.Time

	// compare last seen
	unittest.Equals(t, t1.Equal(t2), false)

	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   2,
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	t3 := result.LastSeen.Time

	// compare last seen
	unittest.Equals(t, t2.Equal(t3), false)

	patrol.shutdown = true
	request = &API_Request{
		ID:    "http",
		Group: "app",
		KeyValue: map[string]interface{}{
			"c": 1,
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, true)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["a"], "b")

	// compare last seen
	unittest.Equals(t, t3.Equal(result.LastSeen.Time), true)

	patrol.shutdown = false
	request = &API_Request{
		ID:    "http",
		Group: "app",
		KeyValue: map[string]interface{}{
			"d": "e",
		},
		KeyValueReplace: true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.KeyValue["a"], "b")
	unittest.Equals(t, result.KeyValue["c"], 1)

	request = &API_Request{
		ID:    "http",
		Group: "app",
		KeyValue: map[string]interface{}{
			"f": "g",
		},
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["d"], "e")

	request = &API_Request{
		ID:    "http",
		Group: "app",
		KeyValue: map[string]interface{}{
			"h": "i",
		},
		KeyValueReplace: true,
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.KeyValue["d"], "e")
	unittest.Equals(t, result.KeyValue["f"], "g")

	request = &API_Request{
		ID:              "http",
		Group:           "app",
		KeyValueReplace: true,
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["h"], "i")

	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// change keep alive
	t4 := result.LastSeen.Time

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_UDP
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t4.Equal(result.LastSeen.Time), false)

	// change keep alive
	t5 := result.LastSeen.Time

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_PID_APP
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t5.Equal(result.LastSeen.Time), false)

	// change keep alive
	t6 := result.LastSeen.Time

	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t6.Equal(result.LastSeen.Time), true)

	// change keep alive
	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_PID_PATROL
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t6.Equal(result.LastSeen.Time), true)

	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.Ping(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t6.Equal(result.LastSeen.Time), true)

	request = &API_Request{
		ID:    "ssh",
		Group: "service",
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "ssh")
	unittest.Equals(t, result.Group, "service")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	request = &API_Request{
		ID:    "ssh",
		Group: "service",
		KeyValue: map[string]interface{}{
			"c": "d",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "ssh")
	unittest.Equals(t, result.Group, "service")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["a"], "b")

	request = &API_Request{
		ID:    "ssh",
		Group: "service",
		KeyValue: map[string]interface{}{
			"e": "f",
		},
		KeyValueReplace: true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "ssh")
	unittest.Equals(t, result.Group, "service")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 2)
	unittest.Equals(t, result.KeyValue["a"], "b")
	unittest.Equals(t, result.KeyValue["c"], "d")

	request = &API_Request{
		ID:              "ssh",
		Group:           "service",
		KeyValueReplace: true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "ssh")
	unittest.Equals(t, result.Group, "service")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["e"], "f")

	request = &API_Request{
		ID:    "ssh",
		Group: "service",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "ssh")
	unittest.Equals(t, result.Group, "service")
	unittest.Equals(t, result.PID, 0)
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.LastSeen.IsZero(), true)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
}
