package patrol

import (
	"fmt"
	"strings"
	"time"
)

type Timestamp struct {
	time.Time
	TimestampFormat string
}

func (self *Timestamp) UnmarshalJSON(
	data []byte,
) error {
	// if our object is nil, we're going to end up using Parse(time.RFC3339) since UnmarshalJSON by default uses this!!!
	f := ""
	if self.TimestampFormat == "" {
		f = time.RFC3339
	} else {
		f = self.TimestampFormat
	}
	s := strings.Trim(string(data), "\"")
	if s == "" || s == "null" {
		self.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(f, s)
	if err != nil {
		return err
	}
	self.Time = t
	return nil
}
func (self Timestamp) MarshalJSON() ([]byte, error) {
	// if our object is nil, we're going to end up using Format(time.RFC3339) since UnmarshalJSON by default uses this!!!
	if self.TimestampFormat == "" {
		return []byte(fmt.Sprintf("\"%s\"", self.Time.Format(time.RFC3339))), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", self.Time.Format(self.TimestampFormat))), nil
}
func (self Timestamp) String() string {
	// if our object is nil, we're going to end up using Format(time.RFC3339) since UnmarshalJSON by default uses this!!!
	if self.TimestampFormat == "" {
		return self.Time.Format(time.RFC3339)
	}
	return self.Time.Format(self.TimestampFormat)
}
