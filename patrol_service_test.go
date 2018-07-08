package patrol

import (
	"log"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolService(t *testing.T) {
	log.Println("TestPatrolService")

	service := &PatrolService{
		Management: 0,
	}
	unittest.Equals(t, service.validate(), ERR_SERVICE_MANAGEMENT_INVALID)
	service.Management = SERVICE_MANAGEMENT_SERVICE

	unittest.Equals(t, service.validate(), ERR_SERVICE_EMPTY)

	service.Service = "1234567890123456790123456789012345678901234567890123456789012345"
	unittest.Equals(t, service.validate(), ERR_SERVICE_MAXLENGTH)

	service.Service = "123456789012345679012345678901234567890123456789012345678901234"

	unittest.Equals(t, service.validate(), ERR_SERVICE_NAME_EMPTY)

	service.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, service.validate(), ERR_SERVICE_NAME_MAXLENGTH)

	service.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	service.IgnoreExitCodes = []uint8{0}
	unittest.Equals(t, service.validate(), ERR_SERVICE_INVALID_EXITCODE)

	service.IgnoreExitCodes = []uint8{1, 1}
	unittest.Equals(t, service.validate(), ERR_SERVICE_DUPLICATE_EXITCODE)

	service.IgnoreExitCodes = []uint8{2}
	unittest.IsNil(t, service.validate())
}
