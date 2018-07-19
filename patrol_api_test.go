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
	// test http listeners
	patrol, err := CreatePatrol(config)
	unittest.Equals(t, err, ERR_LISTEN_HTTP_EMPTY)
	unittest.IsNil(t, patrol)

	// test udp listeners
	config.Apps["http"].KeepAlive = APP_KEEPALIVE_UDP
	patrol, err = CreatePatrol(config)
	unittest.Equals(t, err, ERR_LISTEN_UDP_EMPTY)
	unittest.IsNil(t, patrol)

	// set udp listener
	config.ListenUDP = []string{":0"} // this doesn't matter we aren't using it
	patrol, err = CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)
	// unset
	config.ListenUDP = nil

	// switch back to http
	config.Apps["http"].KeepAlive = APP_KEEPALIVE_HTTP
	config.ListenHTTP = []string{":0"} // this doesn't matter we aren't using it
	patrol, err = CreatePatrol(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, patrol)

	result := patrol.API(nil)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Request NIL")

	request := &API_Request{}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")

	request = &API_Request{
		Group: "?",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Group")

	request = &API_Request{
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")

	request = &API_Request{
		Group: "apps",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown App")

	request = &API_Request{
		Group: "service",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Unknown Service")

	request = &API_Request{
		Group: "services",
		Ping:  true,
	}
	result = patrol.API(request)
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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

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
	unittest.Equals(t, result.PID, 0)
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Toggle: API_TOGGLE_STATE_DISABLE,
		Ping:   true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	// we hadn't previously started our app
	// we pinged to notify our app was "running", so now started is set to last seen
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Toggle: API_TOGGLE_STATE_ENABLE,
		PID:    0,
		Ping:   true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	t1 := result.LastSeen.Time

	// PID no ping
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "PID Requires Ping")

	// ping PID again
	request = &API_Request{
		ID:    "http",
		Group: "app",
		PID:   1,
		Ping:  true,
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)
	t2 := result.LastSeen.Time

	// compare last seen
	unittest.Equals(t, t1.Equal(t2), false)

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
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, true)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["a"], "b")

	// compare last seen
	// lastseen hasn't changed, we didn't ping
	unittest.Equals(t, t2.Equal(result.LastSeen.Time), false)

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
	unittest.Equals(t, result.Started.IsZero(), false)
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
		Ping: true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
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
		Ping:            true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
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
		Ping:            true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 1)
	unittest.Equals(t, result.KeyValue["h"], "i")

	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// change keep alive
	t3 := result.LastSeen.Time

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_UDP
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t3.Equal(result.LastSeen.Time), false)

	// change keep alive
	t4 := result.LastSeen.Time

	// verify ping endpoint
	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_PID_APP
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Ping Not Supported")

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_PID_PATROL
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Ping Not Supported")

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_UDP
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.api(api_endpoint_http, request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Invalid Ping Endpoint")

	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_HTTP
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.api(api_endpoint_udp, request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Invalid Ping Endpoint")

	// valid endpoint
	patrol.apps["http"].config.KeepAlive = APP_KEEPALIVE_HTTP
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.api(api_endpoint_http, request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t4.Equal(result.LastSeen.Time), false)

	// change keep alive
	t5 := result.LastSeen.Time

	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t5.Equal(result.LastSeen.Time), false)

	t6 := result.LastSeen.Time
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// compare last seen
	unittest.Equals(t, t6.Equal(result.LastSeen.Time), false)

	// change PID
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
		PID:   2,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 1)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// copy result
	result_old := result

	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 2)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 1)
	unittest.Equals(t, len(result.KeyValue), 0)
	unittest.Equals(t, result.History[0].PID, 1)
	unittest.Equals(t, result.History[0].Started.Equal(result_old.Started.Time), true)
	unittest.Equals(t, result.History[0].LastSeen.Equal(result_old.LastSeen.Time), true)
	unittest.Equals(t, result.History[0].Stopped.IsZero(), false)
	unittest.Equals(t, result.History[0].Disabled, false)
	unittest.Equals(t, result.History[0].Restart, false)
	unittest.Equals(t, result.History[0].RunOnce, false)
	unittest.Equals(t, result.History[0].Shutdown, false)
	unittest.Equals(t, result.History[0].ExitCode, 0)
	unittest.Equals(t, len(result.History[0].KeyValue), len(result_old.KeyValue))

	// ping again
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
		PID:   2,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 2)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 1)
	unittest.Equals(t, len(result.KeyValue), 0)
	// ping again
	request = &API_Request{
		ID:    "http",
		Group: "app",
		Ping:  true,
		PID:   2,
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.ID, "http")
	unittest.Equals(t, result.Group, "app")
	unittest.Equals(t, result.PID, 2)
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.LastSeen.IsZero(), false)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 1)
	unittest.Equals(t, len(result.KeyValue), 0)

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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
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
	unittest.IsNil(t, result.Started)
	unittest.IsNil(t, result.LastSeen)
	unittest.Equals(t, result.Disabled, false)
	unittest.Equals(t, result.Shutdown, false)
	unittest.Equals(t, len(result.History), 0)
	unittest.Equals(t, len(result.KeyValue), 0)

	// secret test
	// app

	// empty secret
	// this is a GET not a modification
	cas := patrol.apps["http"].GetCAS()
	patrol.apps["http"].config.Secret = "abc"
	request = &API_Request{
		ID:    "http",
		Group: "app",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CASInvalid, true)
	unittest.Equals(t, result.CAS, cas)

	// invalid secret
	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Secret: "ABC",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Secret Invalid")
	unittest.Equals(t, result.CAS, 0)

	// valid secret
	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Secret: "abc",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas)

	// valid secret - keyvalue
	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Secret: "abc",
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas)
	unittest.Equals(t, len(result.KeyValue), 0)

	// valid secret
	request = &API_Request{
		ID:     "http",
		Group:  "app",
		Secret: "abc",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas+1)
	unittest.Equals(t, len(result.KeyValue), 1)

	// service

	// empty secret
	// this is a GET not a modification
	patrol.services["ssh"].config.Secret = "abc"
	cas = patrol.services["ssh"].GetCAS()
	request = &API_Request{
		ID:    "ssh",
		Group: "service",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CASInvalid, true)
	unittest.Equals(t, result.CAS, cas)

	// invalid secret
	request = &API_Request{
		ID:     "ssh",
		Group:  "service",
		Secret: "ABC",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 1)
	unittest.Equals(t, result.Errors[0], "Secret Invalid")
	unittest.Equals(t, result.CAS, 0)

	// valid secret
	request = &API_Request{
		ID:     "ssh",
		Group:  "service",
		Secret: "abc",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas)

	// valid secret - keyvalue
	request = &API_Request{
		ID:     "ssh",
		Group:  "service",
		Secret: "abc",
		KeyValue: map[string]interface{}{
			"a": "b",
		},
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas)
	unittest.Equals(t, len(result.KeyValue), 0)

	// valid secret
	request = &API_Request{
		ID:     "ssh",
		Group:  "service",
		Secret: "abc",
	}
	result = patrol.API(request)
	unittest.Equals(t, len(result.Errors), 0)
	unittest.Equals(t, result.CAS, cas+1)
	unittest.Equals(t, len(result.KeyValue), 1)

}
