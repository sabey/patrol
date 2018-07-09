package patrol

import (
	"fmt"
	"time"
)

type PatrolTimestamp struct {
	time.Time
	f string
}

func (self PatrolTimestamp) MarshalJSON() ([]byte, error) {
	// if our object is nil, we're going to end up using String()
	// if we want to ALWAYS make use of Format, we must set a format even when time is zero
	if self.f == "" {
		return []byte(fmt.Sprintf("\"%s\"", self.Time.String())), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", self.Time.Format(self.f))), nil
}
