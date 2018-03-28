package main

import (
	"fmt"
)

const (
	// service name maximum length in bytes
	SERVICE_NAME_MAXLENGTH = 255
)

const (
	// this is an alias for "service * status"
	// ie: "service ssh status"
	SERVICE_MANAGEMENT_SERVICE = iota + 1
	// this is an alias for "/etc/init.d/* status"
	// ie: "/etc/init.d/ssh status"
	SERVICE_MANAGEMENT_INITD
)

var (
	ERR_SERVICE_NAME_EMPTY         = fmt.Errorf("Service Name was empty")
	ERR_SERVICE_NAME_MAXLENGTH     = fmt.Errorf("Service Name was longer than 255 bytes")
	ERR_SERVICE_MANAGEMENT_INVALID = fmt.Errorf("Service Management was invalid, please select a method!")
	ERR_SERVICE_INVALID_EXITCODE   = fmt.Errorf("Service contained an Invalid Exit Code")
	ERR_SERVICE_DUPLICATE_EXITCODE = fmt.Errorf("Service contained a Duplicate Exit Code")
)

type PatrolService struct {
	// management method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your app may function
	Management int `json:"management,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
	Name string `json:"name,omitempty"`
	// these are a list of valid exit codes to ignore when returned from "service * status"
	// by default 0 is always ignored, it is assumed to mean that the service is running
	IgnoreExitCodes []uint8 `json:"ignore-exit-codes,omitempty"`
}

func (self *PatrolService) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *PatrolService) validate() error {
	if self.Management < SERVICE_MANAGEMENT_SERVICE ||
		self.Management > SERVICE_MANAGEMENT_INITD {
		// unknown management value
		return ERR_SERVICE_MANAGEMENT_INVALID
	}
	if self.Name == "" {
		return ERR_SERVICE_NAME_EMPTY
	}
	if len(self.Name) > SERVICE_NAME_MAXLENGTH {
		return ERR_SERVICE_NAME_MAXLENGTH
	}
	exists := make(map[uint8]struct{})
	for _, ec := range self.IgnoreExitCodes {
		if ec == 0 {
			// can't ignore 0
			return ERR_SERVICE_INVALID_EXITCODE
		}
		if _, ok := exists[ec]; ok {
			// exit code already exists
			return ERR_SERVICE_DUPLICATE_EXITCODE
		}
		// does not exist
		exists[ec] = struct{}{}
	}
	return nil
}
