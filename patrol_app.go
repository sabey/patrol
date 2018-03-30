package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	// app name maximum length in bytes
	APP_NAME_MAXLENGTH = 255
	// ping is used by APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
	APP_PING_EVERY = time.Second * 30
)

// there are multiple methods of process management, none of them are perfect! they all have their tradeoffs!!!
const (
	// KeepAlive is controlled by Patrol:
	// after executing App, Patrol will write the PID to file
	// the upside to this is that we can know exactly when our App exits - we won't have to constantly send "kill -0 PID" to our App to check if it's still alive
	// the downside to this is that if App were to fork and the parent process were to exit we would not be able to track the forked processes, we would then have to respawn our App
	// we're trading fast respawns and external PID access for no child forking
	APP_KEEPALIVE_PID_PATROL = iota + 1
	// KeepAlive is controlled by App:
	// once an App spawns the App is required to write its own PID to file
	// the upside to this is that if the parent process were to fork, the child process would write its new PID to file and Patrol would read and monitor that latest PID instead
	// the downside to this is that we constantly have to load the PID from file and send "kill -0 PID" to check if the process is still alive
	// the other downside is that we may want to optionally check that the PID belongs to AppPath, and if it were not we would have to respawn our App
	// this will allow easy child forking and the ability for the parent process to exit after forking
	// the other trade off is that we won't be able to see the exact time when our monitored PID process exits, leaving a possible delay between respawn
	// see further notes at PatrolApp.PIDVerify
	APP_KEEPALIVE_PID_APP
	// KeepAlive is controlled by HTTP:
	// once an App spawns the App is required intermittently send a HTTP POST to the Patrol JSON API as the keepalive method
	// the upside to this is that we don't have to monitor a PID file and it supports child forking
	// while we won't monitor the PID file here the HTTP POST is required to POST a PID, this way we can send a signal to the PID from the gui admin page for example
	// the downside is that the App must support HTTP clients and have a way to easily do intermittent HTTP POSTs, this might be a burden for most Apps
	APP_KEEPALIVE_HTTP
	// KeepAlive is controlled by UDP:
	// this is similar to HTTP except that it requires less overhead with the added downside that your App does not get a returned confirmation that Patrol is receiving your pings
	// this could be a steep price to pay should the Patrol UDP listener become unresponsive or your UDP packets never arrive, this would result in the App respawning
	APP_KEEPALIVE_UDP
)

var (
	ERR_APP_NAME_EMPTY                  = fmt.Errorf("App Name was empty")
	ERR_APP_NAME_MAXLENGTH              = fmt.Errorf("App Name was longer than 255 bytes")
	ERR_APP_WORKINGDIRECTORY_EMPTY      = fmt.Errorf("App WorkingDirectory was empty")
	ERR_APP_WORKINGDIRECTORY_RELATIVE   = fmt.Errorf("App WorkingDirectory was relative")
	ERR_APP_WORKINGDIRECTORY_UNCLEAN    = fmt.Errorf("App WorkingDirectory was unclean")
	ERR_APP_APPPATH_EMPTY               = fmt.Errorf("App AppPath was empty")
	ERR_APP_APPPATH_UNCLEAN             = fmt.Errorf("App AppPath was unclean")
	ERR_APP_LOGDIRECTORY_EMPTY          = fmt.Errorf("App Log Directory was empty")
	ERR_APP_LOGDIRECTORY_UNCLEAN        = fmt.Errorf("App Log Directory was unclean")
	ERR_APP_KEEPALIVE_INVALID           = fmt.Errorf("App KeepAlive was invalid, please select a keep alive method!")
	ERR_APP_PIDPATH_EMPTY               = fmt.Errorf("App PIDPATH was empty")
	ERR_APP_PIDPATH_UNCLEAN             = fmt.Errorf("App PIDPath was unclean")
	ERR_APP_PIDFILE_NOTFOUND            = fmt.Errorf("App PID File not found")
	ERR_APP_PIDFILE_INVALID             = fmt.Errorf("App PID File was invalid")
	ERR_APP_PING_EXPIRED                = fmt.Errorf("App Ping Expired")
	ERR_APP_KEEPALIVE_PATROL_NOTRUNNING = fmt.Errorf("App KeepAlive Patrol Method not running")
)

type PatrolApp struct {
	// keep alive method
	// this is required! you must choose, it can not default to 0, we can't make assumptions on how your app may function
	KeepAlive int `json:"keepalive,omitempty"`
	// name is only used for the HTTP admin gui, it can contain anything but must be less than 255 bytes in length
	Name string `json:"name,omitempty"`
	// Working Directory is currently required to be non empty
	// we don't want Apps executing relative to the current directory, we want them to know what their reference is
	// IF any other path is relative and not absolute, they will be considered relative to the working directory
	WorkingDirectory string `json:"working-directory,omitempty"`
	// App Path to the app executable
	AppPath string `json:"app-path,omitempty"`
	// Log Directory for stderr and stdout
	LogDirectory string `json:"log-directory,omitempty"`
	// path to pid file
	// PID is optional, it is only required when using the PATROL or APP keepalive methods
	PIDPath string `json:"pid-path,omitempty"`
	// should we verify that the PID belongs to AppPath?
	// the reason for this is that it is technically possible for your App to write a PID to file, exit, and then for another long running service to start with this same PID
	// the problem here is that that other long running process would be confused for our App and we would assume it is running
	// the only solution is to verify the processes name OR for you to continuously write your PID to file in intervals, say write to PID every 10 seconds
	// the problem with the constant PID writing is that should your parent fork and create a child, you would want to stop writing the parent PID and only write the child PID!
	PIDVerify bool `json:"pid-verify,omitempty"`
	// private
	// this is only used by APP_KEEPALIVE_PID_PATROL
	// this should almost always be true, since we're handling our process management ourselves
	// the only reason this should be false is if we're between reexecuting or we've stopped our process
	is_running bool
	// pid is set by all but only used by APP_KEEPALIVE_PID_APP to verify a process is running
	// last PID we've found and verified
	// the maximum PID value on a 32 bit system is 32767
	// the maximum PID value on a 64 bit system is 2^22
	// systems by default will default to 32767, but we should support up to uint32
	pid uint32
	// ping is used by APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
	ping time.Time
	mu   sync.RWMutex
}

func (self *PatrolApp) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *PatrolApp) validate() error {
	if self.KeepAlive < APP_KEEPALIVE_PID_PATROL ||
		self.KeepAlive > APP_KEEPALIVE_UDP {
		// unknown keep alive value
		return ERR_APP_KEEPALIVE_INVALID
	}
	if self.Name == "" {
		return ERR_APP_NAME_EMPTY
	}
	if len(self.Name) > APP_NAME_MAXLENGTH {
		return ERR_APP_NAME_MAXLENGTH
	}
	if self.WorkingDirectory == "" {
		return ERR_APP_WORKINGDIRECTORY_EMPTY
	}
	if self.WorkingDirectory[0] != '/' {
		// working directory can not be relative
		// we require that it is absolute, so that other paths may be relative to it
		return ERR_APP_WORKINGDIRECTORY_RELATIVE
	}
	if !IsPathClean(self.WorkingDirectory) {
		return ERR_APP_WORKINGDIRECTORY_UNCLEAN
	}
	if self.AppPath == "" {
		return ERR_APP_APPPATH_EMPTY
	}
	if !IsPathClean(self.AppPath) {
		return ERR_APP_APPPATH_UNCLEAN
	}
	if self.LogDirectory == "" {
		return ERR_APP_LOGDIRECTORY_EMPTY
	}
	if !IsPathClean(self.LogDirectory) {
		return ERR_APP_LOGDIRECTORY_UNCLEAN
	}
	if self.KeepAlive == APP_KEEPALIVE_PID_PATROL ||
		self.KeepAlive == APP_KEEPALIVE_PID_APP {
		// PID is required
		if self.PIDPath == "" {
			return ERR_APP_PIDPATH_EMPTY
		}
		if !IsPathClean(self.PIDPath) {
			return ERR_APP_PIDPATH_UNCLEAN
		}
	}
	return nil
}
func (self *PatrolApp) startApp() error {
	// we can't set WorkingDirectory and only execute just AppPath
	// we must use the absolute path of WorkingDirectory and AppPath for execute to work properly
	cmd := exec.Command(filepath.Clean(self.WorkingDirectory + "/" + self.AppPath))
	// we still have to set our WorkingDirectory
	cmd.Dir = self.WorkingDirectory
	// SysProcAttr holds optional, operating system-specific attributes.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// we want our children to have their own process group IDs
		// the reason for this is that we want them to run on their own
		// our shell by default will act as a catchall for the signals
		// ideally we would like our children to receive their own signal
		Setpgid: true,
		// 0 causes us to set the group id to the process id
		Pgid: 0,
		// Signal that the process will get when its parent dies (Linux only)
		// we can't rely on this and if were able to we should notify our children
		//
		// The SIGTERM signal is a generic signal used to cause program termination.
		// This signal can be blocked, handled, and ignored. It is the normal way to politely ask a program to terminate
		// we don't have to close our process, but we should be aware that we're not being monitored
		// some processes may notice they receive 2 SIGTERMS, I'm not sure why it's doing this, just ignore additional signals
		Pdeathsig: syscall.SIGTERM,
	}
	if unittesting {
		// we're going to hijack our stderr and stdout for easy debugging
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	// start will start our process but will not wait for execute to finish running
	if err := cmd.Start(); err != nil {
		// failed to start
		return err
	}
	// we have to call Wait() on our process and read the exit code
	// if we don't we will end up with a zombie process
	// zombie processes don't use a lot of system resources, but they will retain their PID
	// we're just going to discard this action, we don't care what the exit code is, ideally later we can log this code in history
	// as of right now for APP_KEEPALIVE_PID_APP we don't always expect to see an exit code as we're expecting children to fork
	// tracking of the exit code makes a lot of sense for APP_KEEPALIVE_PID_PATROL because we ALWAYS see the exit code
	go cmd.Wait()
	return nil
}
func (self *PatrolApp) isAppRunning() error {
	if self.KeepAlive == APP_KEEPALIVE_HTTP ||
		self.KeepAlive == APP_KEEPALIVE_UDP {
		// check if we've been pinged recently
		self.mu.RLock()
		ping := self.ping
		self.mu.RUnlock()
		// if last ping + ping timeout is NOT after now we know that we've timedout
		if time.Now().After(ping.Add(APP_PING_EVERY)) {
			// expired
			return ERR_APP_PING_EXPIRED
		}
		// still alive
		return nil
	} else if self.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// check our internal state
		self.mu.RLock()
		defer self.mu.RUnlock()
		if !self.is_running {
			// not running
			return ERR_APP_KEEPALIVE_PATROL_NOTRUNNING
		}
		// running
		return nil
	}
	// we have to ping our PID to determine if we're running
	// this function is only used by APP_KEEPALIVE_PID_APP
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	pid, err := self.getPID()
	if err != nil {
		// failed to find PID
		return err
	}
	// TODO: we should add PID verification here
	// either before or after we signal to kill, it's unsure how this will work
	cmd := exec.CommandContext(ctx, "kill", "-0", fmt.Sprintf("%d", pid))
	return cmd.Run()
}
func (self *PatrolApp) getPID() (
	uint32,
	error,
) {
	// this function is only used by APP_KEEPALIVE_PID_APP
	// we must use the absolute path of our WorkingDirectory and AppPath to find our PID
	file, err := os.Open(filepath.Clean(self.WorkingDirectory + "/" + self.PIDPath))
	if err != nil {
		// failed to open PID
		return 0, ERR_APP_PIDFILE_NOTFOUND
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		// pid was invalid
		return 0, ERR_APP_PIDFILE_INVALID
	}
	pid, err := strconv.ParseUint(string(bytes.TrimSpace(b)), 10, 16)
	if err != nil {
		// failed to parse PID
		return 0, ERR_APP_PIDFILE_INVALID
	}
	self.mu.Lock()
	self.pid = uint32(pid)
	self.mu.Unlock()
	return uint32(pid), nil
}
func (self *PatrolApp) GetPID() uint32 {
	// this may not be the latest PID but it's the latest PID we're aware of
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.pid
}
