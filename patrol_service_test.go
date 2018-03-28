package main

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
