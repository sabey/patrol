package patrol

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

type Service struct {
	// safe
	config *ConfigService
	// unsafe
}

func (self *Service) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Service) GetConfig() *ConfigService {
	return self.config.Clone()
}
func (self *Service) startService() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	var cmd *exec.Cmd
	if self.config.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, "start")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), "start")
	}
	return cmd.Run()
}
func (self *Service) isServiceRunning() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	var cmd *exec.Cmd
	if self.config.Management == SERVICE_MANAGEMENT_SERVICE {
		cmd = exec.CommandContext(ctx, "service", self.config.Service, "status")
	} else {
		cmd = exec.CommandContext(ctx, fmt.Sprintf("/etc/init.d/%s", self.config.Service), "status")
	}
	return cmd.Run()
}
