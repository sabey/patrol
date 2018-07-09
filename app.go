package patrol

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// app name maximum length in bytes
	APP_NAME_MAXLENGTH = 255
	// ping is used by APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
	APP_PING_EVERY = time.Second * 30
	// environment keys
	APP_ENV_KEEPALIVE    = `PATROL_KEEPALIVE`
	APP_ENV_PID_PATH     = `PATROL_PID`
	APP_ENV_HTTP_ADDRESS = `PATROL_HTTP_ADDRESS`
	APP_ENV_UDP_ADDRESS  = `PATROL_UDP_ADDRESS`
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
	// the other downside is that we may want to optionally check that the PID belongs to Binary, and if it were not we would have to respawn our App
	// this will allow easy child forking and the ability for the parent process to exit after forking
	// the other trade off is that we won't be able to see the exact time when our monitored PID process exits, leaving a possible delay between respawn
	// see further notes at App.PIDVerify
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
	ERR_APP_PING_EXPIRED                = fmt.Errorf("App Ping Expired")
	ERR_APP_KEEPALIVE_PATROL_NOTRUNNING = fmt.Errorf("App KeepAlive Patrol Method not running")
	ERR_APP_PIDFILE_NOTFOUND            = fmt.Errorf("App PID File not found")
	ERR_APP_PIDFILE_INVALID             = fmt.Errorf("App PID File was invalid")
)

type App struct {
	// safe
	patrol *Patrol
	id     string // we want a reference to our parent ID
	config *ConfigApp
	// unsafe
	keyvalue map[string]interface{}
	history  []*History
	started  time.Time
	lastseen time.Time
	disabled bool // takes its initial value from config
	// pid is set by all but only used by APP_KEEPALIVE_PID_APP to verify a process is running
	// last PID we've found and verified
	// the maximum PID value on a 32 bit system is 32767
	// the maximum PID value on a 64 bit system is 2^22
	// systems by default will default to 32767, but we should support up to uint32
	pid uint32
	// we're going to save our exit code for history
	// this is only supported by APP_KEEPALIVE_PID_PATROL
	exit_code uint8
	// ping is used by APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
	ping time.Time
	mu   sync.RWMutex
}

func (self *App) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *App) GetID() string {
	return self.id
}
func (self *App) GetPatrol() *Patrol {
	return self.patrol
}
func (self *App) GetConfig() *ConfigApp {
	return self.config.Clone()
}
func (self *App) IsRunning() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return !self.started.IsZero()
}
func (self *App) IsDisabled() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.disabled
}
func (self *App) Disable() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.disabled = true
}
func (self *App) Enable() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.disabled = false
}
func (self *App) GetStarted() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.started
}
func (self *App) GetLastSeen() time.Time {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.lastseen
}
func (self *App) GetKeyValue() map[string]interface{} {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.getKeyValue()
}
func (self *App) getKeyValue() map[string]interface{} {
	// dereference
	kv := make(map[string]interface{})
	for k, v := range self.keyvalue {
		kv[k] = v
	}
	return kv
}
func (self *App) SetKeyValue(
	kv map[string]interface{},
) {
	self.mu.Lock()
	for k, v := range kv {
		self.keyvalue[k] = v
	}
	self.mu.Unlock()
}
func (self *App) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.mu.Lock()
	self.keyvalue = make(map[string]interface{})
	// dereference
	for k, v := range kv {
		self.keyvalue[k] = v
	}
	self.mu.Unlock()
}
func (self *App) GetHistory() []*History {
	self.mu.RLock()
	defer self.mu.RUnlock()
	// dereference
	history := make([]*History, 0, len(self.history))
	for _, h := range self.history {
		history = append(history, h.clone())
	}
	return history
}
func (self *App) close() {
	if !self.started.IsZero() {
		// save history
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			// we're always going to log PID even if there's a chance it doesn't exist
			// for example if our APP controls the PID, when we ping to check if its alive, it would override PID with something incorrect
			// pid is garaunteed to always exist for APP_KEEPALIVE_PID_PATROL
			PID: self.pid,
			Started: PatrolTimestamp{
				Time: self.started,
				f:    self.patrol.config.Timestamp,
			},
			LastSeen: PatrolTimestamp{
				Time: self.lastseen,
				f:    self.patrol.config.Timestamp,
			},
			Stopped: PatrolTimestamp{
				Time: time.Now(),
				f:    self.patrol.config.Timestamp,
			},
			Disabled: self.disabled,
			Shutdown: self.patrol.shutdown,
			// exit code is only garaunteed to exist for APP_KEEPALIVE_PID_PATROL
			ExitCode: self.exit_code,
			KeyValue: self.getKeyValue(),
		}
		self.history = append(self.history, h)
		// reset values
		self.started = time.Time{}
		self.pid = 0
		self.exit_code = 0
		if self.config.KeyValueClear {
			// clear keyvalues
			self.keyvalue = make(map[string]interface{})
		}
		// call trigger in a go routine so we don't deadlock
		if self.config.TriggerClosed != nil {
			go self.config.TriggerClosed(self.id, self, h)
		}
	}
}
func (self *App) startApp() error {
	// close previous if it exists
	self.close()
	now := time.Now()
	// we can't set WorkingDirectory and only execute just Binary
	// we must use the absolute path of WorkingDirectory and Binary for execute to work properly
	var cmd *exec.Cmd
	if self.config.ExecuteTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(self.config.ExecuteTimeout))
		defer cancel()
		cmd = exec.CommandContext(ctx, filepath.Clean(self.config.WorkingDirectory+"/"+self.config.Binary))
	} else {
		cmd = exec.Command(filepath.Clean(self.config.WorkingDirectory + "/" + self.config.Binary))
	}
	// Args
	if len(self.config.Args) > 0 {
		cmd.Args = self.config.Args
	}
	if self.config.ExtraArgs != nil {
		if a := self.config.ExtraArgs(self.id, self); len(a) > 0 {
			cmd.Args = append(cmd.Args, a...)
		}
	}
	// Env
	// we're going to include our own environment variables
	// so EnvParent would be important, since if a user expects nil Env they'll never get parent variables
	if self.config.EnvParent {
		// include parents environment variables
		cmd.Env = os.Environ()
	}
	if len(self.config.Env) > 0 {
		cmd.Env = append(cmd.Env, self.config.Env...)
	}
	if self.config.ExtraEnv != nil {
		if e := self.config.ExtraEnv(self.id, self); len(e) > 0 {
			cmd.Env = append(cmd.Env, e...)
		}
	}
	// patrol environment variables
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, self.config.KeepAlive))
	if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL ||
		self.config.KeepAlive == APP_KEEPALIVE_PID_APP {
		// pid path
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", APP_ENV_PID_PATH, filepath.Clean(self.config.WorkingDirectory+"/"+self.config.PIDPath)))
	} else if self.config.KeepAlive == APP_KEEPALIVE_HTTP {
		// http address
	} else if self.config.KeepAlive == APP_KEEPALIVE_UDP {
		// udp address
	}
	// STD in/out/err
	if self.config.Stdin != nil {
		cmd.Stdin = self.config.Stdin
	}
	if self.config.Stdout != nil {
		cmd.Stdout = self.config.Stdout
	} else {
		// we need to create a stdout file
		// create an ordered log directory
		// this is the exact same as `mkdir -p`
		ld := self.logDir()
		err := os.MkdirAll(ld, os.ModePerm)
		if err != nil {
			log.Printf("./patrol.startApp(): App ID: %s Stdout failed to MkdirAll: \"%s\" Err: \"%s\"\n", self.id, ld, err)
			return err
		}
		// use now as our unique key
		fn := fmt.Sprintf("%s/%d.stdout.log", ld, now.UnixNano())
		cmd.Stdout, err = OpenFile(fn)
		if err != nil {
			log.Printf("./patrol.startApp(): App ID: %s Stdout failed to OpenFile: \"%s\" Err: \"%s\"\n", self.id, fn, err)
			return err
		}
		// we CAN NOT defer close this file!!!
		// we are passing this file handler to the app we are executing
		// our executed app will handle closing this file descriptor on close
	}
	if self.config.Stderr != nil {
		cmd.Stderr = self.config.Stderr
	} else {
		// we need to create a stderr file
		// create an ordered log directory
		// this is the exact same as `mkdir -p`
		ld := self.logDir()
		err := os.MkdirAll(ld, os.ModePerm)
		if err != nil {
			log.Printf("./patrol.startApp(): App ID: %s Stderr failed to MkdirAll: \"%s\" Err: \"%s\"\n", self.id, ld, err)
			return err
		}
		// use now as our unique key
		fn := fmt.Sprintf("%s/%d.stderr.log", ld, now.UnixNano())
		cmd.Stderr, err = OpenFile(fn)
		if err != nil {
			log.Printf("./patrol.startApp(): App ID: %s Stderr failed to OpenFile: \"%s\" Err: \"%s\"\n", self.id, fn, err)
			return err
		}
		// we CAN NOT defer close this file!!!
		// we are passing this file handler to the app we are executing
		// our executed app will handle closing this file descriptor on close
	}
	// extra files
	if self.config.ExtraFiles != nil {
		if e := self.config.ExtraFiles(self.id, self); len(e) > 0 {
			cmd.ExtraFiles = e
		}
	}
	// we still have to set our WorkingDirectory
	cmd.Dir = self.config.WorkingDirectory
	// SysProcAttr holds optional, operating system-specific attributes.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// we want our children to have their own process group IDs
		// the reason for this is that we want them to run on their own
		// our shell by default which will act as a catchall for the signals
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
	// start will start our process but will not wait for execute to finish running
	if err := cmd.Start(); err != nil {
		// failed to start
		return err
	}
	// started!
	self.started = now
	if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// we're going to copy our PID from our process
		// any other keep alive method we're just going to ignore the process PID and assume it's wrong
		self.pid = uint32(cmd.Process.Pid)
	}
	// we have to call Wait() on our process and read the exit code
	// if we don't we will end up with a zombie process
	// zombie processes don't use a lot of system resources, but they will retain their PID
	// we're just going to discard this action, we don't care what the exit code is, ideally later we can log this code in history
	// as of right now for APP_KEEPALIVE_PID_APP we don't always expect to see an exit code as we're expecting children to fork
	// tracking of the exit code makes a lot of sense for APP_KEEPALIVE_PID_PATROL because we ALWAYS see the exit code
	go func() {
		// we're going to need add functionality for if we choose to signal this command to stop
		// we can either wrap our context? or use os.Process.Kill
		// ideally we would want to use our context, because we're not sure what we would be signalling a kill to if this stopped before the kill
		// context seems like the most ideal path to choose
		err := cmd.Wait()
		var exit_code uint8 = 0
		if err != nil && self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
			// we're going to copy our exit code from our result
			// any other keep alive method we're just going to ignore the exit code and assume it's wrong
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0
				// This works on both Unix and Windows.
				// Although package syscall is generally platform dependent,
				// WaitStatus is defined for both Unix and Windows and in both cases has an ExitStatus() method with the same signature.
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exit_code = uint8(status.ExitStatus())
				}
			}
		}
		// currently this can't race because we ALWAYS check isAppRunning() before startApp() AND we only use tick() to start services
		// this logic should never change, so it's not something to worry about right now
		self.mu.Lock()
		// set exit code
		self.exit_code = exit_code
		// close app
		self.close()
		self.mu.Unlock()
	}()
	return nil
}
func (self *App) isAppRunning() error {
	// check
	if self.config.KeepAlive == APP_KEEPALIVE_HTTP ||
		self.config.KeepAlive == APP_KEEPALIVE_UDP {
		// check if we've been pinged recently
		// if last ping + ping timeout is NOT after now we know that we've timedout
		if time.Now().After(self.ping.Add(APP_PING_EVERY)) {
			// expired
			// close app
			self.close()
			return ERR_APP_PING_EXPIRED
		}
		// running!
		return nil
	} else if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// check our internal state
		if self.started.IsZero() {
			// not running
			// we do NOT have to save history!!!
			// our teardown function after cmd.Wait() will save our history!
			return ERR_APP_KEEPALIVE_PATROL_NOTRUNNING
		}
		// running!
		return nil
	}
	// we have to ping our PID to determine if we're running
	// this function is only used by APP_KEEPALIVE_PID_APP
	pid, err := self.getPID()
	if err != nil {
		// failed to find PID
		// close app
		self.close()
		return err
	}
	// TODO: we should add PID verification here
	// either before or after we signal to kill, it's unsure how this will work
	process, err := os.FindProcess(int(pid))
	if err != nil {
		// NOT running!
		// close app
		self.close()
		return err
	}
	// kill -0 PID
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// NOT running!
		// close app
		self.close()
		return err
	}
	// running!
	now := time.Now()
	if self.started.IsZero() {
		// we need to set started since this is our first time seeing this app
		self.started = now
	}
	self.lastseen = now
	return nil
}
func (self *App) signalStop() {
	// we're going to signal our App if our App is disabled OR if we're shutting down Patrol
	// we can only do this if we have a PID, we don't care what keepalive method we use so long as a PID exists
	// we're going to discard any errors
	if self.pid > 0 {
		if process, err := os.FindProcess(int(self.pid)); err == nil {
			// we're going to keep our signals different than syscall.SIGTERM
			// we're going to leave syscall.SIGTERM to be reserved for Patrol ACTUALLY closing!
			if self.patrol.shutdown {
				// notify that Patrol is gracefully shutting down
				process.Signal(syscall.SIGUSR1)
			} else {
				// notify that the App has been disabled
				process.Signal(syscall.SIGUSR2)
			}
		}
	}
}
func (self *App) getPID() (
	uint32,
	error,
) {
	// this function is only used by APP_KEEPALIVE_PID_APP
	// we must use the absolute path of our WorkingDirectory and Binary to find our PID
	file, err := os.Open(filepath.Clean(self.config.WorkingDirectory + "/" + self.config.PIDPath))
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
	self.pid = uint32(pid)
	return self.pid, nil
}
func (self *App) GetPID() uint32 {
	// this may not be the latest PID but it's the latest PID we're aware of
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.pid
}
func (self *App) logDir() string {
	y, m, d := time.Now().Date()
	return filepath.Clean(
		fmt.Sprintf(
			"%s/%s/%d/%s/%d",
			self.config.WorkingDirectory,
			self.config.LogDirectory,
			y,
			strings.ToLower(m.String()),
			d,
		),
	)
}
