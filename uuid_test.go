package patrol

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestUUID(t *testing.T) {
	log.Println("TestUUID")

	uuid, err := uuidV4()
	unittest.IsNil(t, err)
	unittest.Equals(t, uuid != "", true)

	for i := 0; i < 100; i++ {
		uuid = uuidMust(uuidV4())
		// xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx
		for i := 0; i < 36; i++ {
			if i == 8 ||
				i == 13 ||
				i == 18 ||
				i == 23 {
				unittest.Equals(t, uuid[i], '-')
			} else {
				unittest.Equals(t, uuid[i] != '-', true)
			}
		}
	}
}
