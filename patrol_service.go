package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const (
	// service name maximum length in bytes
	SERVICE_NAME_MAXLENGTH = 255
	// service maximum length in bytes
	SERVICE_MAXLENGTH = 63
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
	ERR_SERVICE_EMPTY              = fmt.Errorf("Service was empty")
	ERR_SERVICE_MAXLENGTH          = fmt.Errorf("Service was longer than 63 bytes")
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
	// Service is the name of the executable and/or parameter
	Service string `json:"service,omitempty"`
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
	if self.Service == "" {
		return ERR_SERVICE_EMPTY
	}
	if len(self.Service) > SERVICE_MAXLENGTH {
		return ERR_SERVICE_MAXLENGTH
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
func (self *PatrolService) startService() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	if self.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.Service, "start")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.Service), "start")
	}
	return cmd.Run()
}
func (self *PatrolService) isServiceRunning() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	var cmd *exec.Cmd
	if self.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.Service, "status")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.Service), "status")
	}
	return cmd.Run()
}
