package patrol

import (
	"encoding/json"
	"log"
	"sabey.co/unittest"
	"testing"
	"time"
)

func TestAPIResponse(t *testing.T) {
	log.Println("TestAPIResponse")

	now := time.Now()

	log.Println("marshal")
	response := &API_Response{
		ID:    "id",
		Group: "group",
		Name:  "name",
		PID:   1,
		Started: &Timestamp{
			Time: now,
		},
		LastSeen: &Timestamp{
			Time: now,
		},
		Disabled: true,
		Restart:  true,
		RunOnce:  true,
		Shutdown: true,
		History: []*History{
			&History{
				PID: 1,
				Started: &Timestamp{
					Time: now,
				},
				LastSeen: &Timestamp{
					Time: now,
				},
				Stopped: &Timestamp{
					Time: now,
				},
				Disabled: true,
				Restart:  true,
				RunOnce:  true,
				Shutdown: true,
				ExitCode: 1,
				KeyValue: map[string]interface{}{
					"a": "b",
				},
			},
		},
		KeyValue: map[string]interface{}{
			"a": "b",
		},
		Secret: true,
		Errors: []string{
			"a",
		},
		CAS:        1,
		CASInvalid: true,
	}
	body, err := json.MarshalIndent(response, "", "\t")
	unittest.IsNil(t, err)
	unittest.Equals(t, len(body) > 0, true)
	log.Printf("response: \"%s\"\n", body)

	log.Println("unmarshal")
	result := &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// response
	unittest.Equals(t, response.ID, result.ID)
	unittest.Equals(t, response.Group, result.Group)
	unittest.Equals(t, response.Name, result.Name)
	unittest.Equals(t, response.PID, result.PID)
	unittest.Equals(t, response.Started.String(), result.Started.String())   // we can't use Equal due to response having being a monotonic timestamp
	unittest.Equals(t, response.LastSeen.String(), result.LastSeen.String()) // we can't use Equal due to response having being a monotonic timestamp
	unittest.Equals(t, response.Disabled, result.Disabled)
	unittest.Equals(t, response.Restart, result.Restart)
	unittest.Equals(t, response.RunOnce, result.RunOnce)
	unittest.Equals(t, response.Shutdown, result.Shutdown)
	unittest.Equals(t, len(response.History), len(result.History))
	unittest.Equals(t, len(response.KeyValue), len(result.KeyValue))
	unittest.Equals(t, response.KeyValue["a"], result.KeyValue["a"])
	unittest.Equals(t, response.Secret, result.Secret)
	unittest.Equals(t, len(response.Errors), len(result.Errors))
	unittest.Equals(t, response.Errors[0], result.Errors[0])
	unittest.Equals(t, response.CAS, result.CAS)
	unittest.Equals(t, response.CASInvalid, result.CASInvalid)
	// history
	unittest.Equals(t, response.History[0].PID, result.History[0].PID)
	unittest.Equals(t, response.History[0].Started.String(), result.History[0].Started.String())
	unittest.Equals(t, response.History[0].LastSeen.String(), result.History[0].LastSeen.String())
	unittest.Equals(t, response.History[0].Stopped.String(), result.History[0].Stopped.String())
	unittest.Equals(t, response.History[0].Disabled, result.History[0].Disabled)
	unittest.Equals(t, response.History[0].Restart, result.History[0].Restart)
	unittest.Equals(t, response.History[0].RunOnce, result.History[0].RunOnce)
	unittest.Equals(t, response.History[0].Shutdown, result.History[0].Shutdown)
	unittest.Equals(t, response.History[0].ExitCode, result.History[0].ExitCode)
	unittest.Equals(t, response.History[0].PID, result.History[0].PID)
	unittest.Equals(t, len(response.History[0].KeyValue), len(result.History[0].KeyValue))
	unittest.Equals(t, response.History[0].KeyValue["a"], result.History[0].KeyValue["a"])
}
func TestAPIResponseCustom(t *testing.T) {
	log.Println("TestAPIResponseCustom")

	now := time.Now()

	log.Println("marshal")
	response := &API_Response{
		ID:    "id",
		Group: "group",
		Name:  "name",
		PID:   1,
		Started: &Timestamp{
			Time:            now,
			TimestampFormat: time.UnixDate,
		},
		LastSeen: &Timestamp{
			Time:            now,
			TimestampFormat: time.UnixDate,
		},
		Disabled: true,
		Restart:  true,
		RunOnce:  true,
		Shutdown: true,
		History: []*History{
			&History{
				PID: 1,
				Started: &Timestamp{
					Time:            now,
					TimestampFormat: time.UnixDate,
				},
				LastSeen: &Timestamp{
					Time:            now,
					TimestampFormat: time.UnixDate,
				},
				Stopped: &Timestamp{
					Time:            now,
					TimestampFormat: time.UnixDate,
				},
				Disabled: true,
				Restart:  true,
				RunOnce:  true,
				Shutdown: true,
				ExitCode: 1,
				KeyValue: map[string]interface{}{
					"a": "b",
				},
			},
		},
		KeyValue: map[string]interface{}{
			"a": "b",
		},
		Secret: true,
		Errors: []string{
			"a",
		},
		CAS:        1,
		CASInvalid: true,
	}
	body, err := json.MarshalIndent(response, "", "\t")
	unittest.IsNil(t, err)
	unittest.Equals(t, len(body) > 0, true)
	log.Printf("response: \"%s\"\n", body)

	log.Println("unmarshal")
	result := &API_Response{
		patrol: &Patrol{
			config: &Config{
				Timestamp: time.UnixDate,
			},
		},
	}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// response
	unittest.Equals(t, response.ID, result.ID)
	unittest.Equals(t, response.Group, result.Group)
	unittest.Equals(t, response.Name, result.Name)
	unittest.Equals(t, response.PID, result.PID)
	unittest.Equals(t, response.Started.String(), result.Started.String())   // we can't use Equal due to response having being a monotonic timestamp
	unittest.Equals(t, response.LastSeen.String(), result.LastSeen.String()) // we can't use Equal due to response having being a monotonic timestamp
	unittest.Equals(t, response.Disabled, result.Disabled)
	unittest.Equals(t, response.Restart, result.Restart)
	unittest.Equals(t, response.RunOnce, result.RunOnce)
	unittest.Equals(t, response.Shutdown, result.Shutdown)
	unittest.Equals(t, len(response.History), len(result.History))
	unittest.Equals(t, len(response.KeyValue), len(result.KeyValue))
	unittest.Equals(t, response.KeyValue["a"], result.KeyValue["a"])
	unittest.Equals(t, response.Secret, result.Secret)
	unittest.Equals(t, len(response.Errors), len(result.Errors))
	unittest.Equals(t, response.Errors[0], result.Errors[0])
	unittest.Equals(t, response.CAS, result.CAS)
	unittest.Equals(t, response.CASInvalid, result.CASInvalid)
	// history
	unittest.Equals(t, response.History[0].PID, result.History[0].PID)
	unittest.Equals(t, response.History[0].Started.String(), result.History[0].Started.String())
	unittest.Equals(t, response.History[0].LastSeen.String(), result.History[0].LastSeen.String())
	unittest.Equals(t, response.History[0].Stopped.String(), result.History[0].Stopped.String())
	unittest.Equals(t, response.History[0].Disabled, result.History[0].Disabled)
	unittest.Equals(t, response.History[0].Restart, result.History[0].Restart)
	unittest.Equals(t, response.History[0].RunOnce, result.History[0].RunOnce)
	unittest.Equals(t, response.History[0].Shutdown, result.History[0].Shutdown)
	unittest.Equals(t, response.History[0].ExitCode, result.History[0].ExitCode)
	unittest.Equals(t, response.History[0].PID, result.History[0].PID)
	unittest.Equals(t, len(response.History[0].KeyValue), len(result.History[0].KeyValue))
	unittest.Equals(t, response.History[0].KeyValue["a"], result.History[0].KeyValue["a"])
}
