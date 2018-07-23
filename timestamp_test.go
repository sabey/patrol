package patrol

import (
	"encoding/json"
	"log"
	"sabey.co/unittest"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	log.Println("TestTimestamp")

	ts := &Timestamp{}
	unittest.IsNil(t, json.Unmarshal([]byte("\"\""), ts))
	unittest.Equals(t, ts.IsZero(), true)

	ts = &Timestamp{}
	unittest.IsNil(t, json.Unmarshal([]byte("\"null\""), ts))
	unittest.Equals(t, ts.IsZero(), true)

	ts = &Timestamp{}
	unittest.IsNil(t, json.Unmarshal([]byte("null"), ts))
	unittest.Equals(t, ts.IsZero(), true)

	ts = &Timestamp{}
	now := time.Now()
	unittest.IsNil(t, json.Unmarshal([]byte("\""+now.Format(time.RFC3339)+"\""), ts))
	unittest.Equals(t, ts.Time.Format(time.RFC3339), now.Format(time.RFC3339))

	ts = &Timestamp{
		TimestampFormat: time.UnixDate,
	}
	now = time.Now()
	unittest.IsNil(t, json.Unmarshal([]byte("\""+now.Format(time.UnixDate)+"\""), ts))
	unittest.Equals(t, ts.Time.Format(time.UnixDate), now.Format(time.UnixDate))
}
