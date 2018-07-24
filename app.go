package patrol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sabey.co/patrol/cas"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// app name maximum length in bytes
	APP_NAME_MAXLENGTH = 255
	// ping is used by APP_KEEPALIVE_HTTP and APP_KEEPALIVE_UDP
	APP_PING_TIMEOUT_MIN     = 5
	APP_PING_TIMEOUT_DEFAULT = 30
	APP_PING_TIMEOUT_MAX     = 180
	// environment keys
	APP_ENV_APP_ID      = `PATROL_ID`
	APP_ENV_KEEPALIVE   = `PATROL_KEEPALIVE`
	APP_ENV_PID         = `PATROL_PID`
	APP_ENV_LISTEN_HTTP = `PATROL_HTTP`
	APP_ENV_LISTEN_UDP  = `PATROL_UDP`
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
	// history will wrap our cas Objects Lock/RLock mutex
	// history is NOT included in our cas Object because we didn't want to restructure Patrol
	history []*History
	o       *cas.App
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
func (self *App) GetCAS() uint64 {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetCAS()
}
func (self *App) IsRunning() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return !self.o.GetStarted().IsZero()
}
func (self *App) GetStarted() time.Time {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetStarted()
}
func (self *App) GetStartedLog() time.Time {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetStartedLog()
}
func (self *App) GetLastSeen() time.Time {
	self.o.RLock()
	defer self.o.RUnlock()
	if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// if our app is running lastseen should exist
		if !self.o.GetStarted().IsZero() {
			// we're running
			return time.Now()
		}
	}
	return self.o.GetLastSeen()
}
func (self *App) IsDisabled() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsDisabled()
}
func (self *App) IsRestart() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRestart()
}
func (self *App) IsRunOnce() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRunOnce()
}
func (self *App) IsRunOnceConsumed() bool {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.IsRunOnceConsumed()
}
func (self *App) Toggle(
	toggle uint8,
) {
	self.o.Lock()
	self.toggle(toggle)
	self.o.Unlock()
}
func (self *App) Enable() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_ENABLE)
	self.o.Unlock()
}
func (self *App) Disable() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_DISABLE)
	self.o.Unlock()
}
func (self *App) Restart() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RESTART)
	self.o.Unlock()
}
func (self *App) EnableRunOnce() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_ENABLE)
	self.o.Unlock()
}
func (self *App) DisableRunOnce() {
	self.o.Lock()
	self.toggle(API_TOGGLE_STATE_RUNONCE_DISABLE)
	self.o.Unlock()
}
func (self *App) GetKeyValue() map[string]interface{} {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetKeyValue()
}
func (self *App) SetKeyValue(
	kv map[string]interface{},
) {
	self.o.Lock()
	self.o.SetKeyValue(kv)
	self.o.Unlock()
}
func (self *App) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.o.Lock()
	self.o.ReplaceKeyValue(kv)
	self.o.Unlock()
}
func (self *App) GetHistory() []*History {
	self.o.RLock()
	defer self.o.RUnlock()
	return self.getHistory()
}
func (self *App) getHistory() []*History {
	// dereference
	history := make([]*History, 0, len(self.history))
	for _, h := range self.history {
		history = append(history, h.clone())
	}
	return history
}
func (self *App) close() {
	if !self.o.GetStarted().IsZero() {
		now := time.Now()
		// save history
		self.o.Increment() // we have to increment for modifying History
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			// we're always going to log PID even if there's a chance it doesn't exist
			// for example if our APP controls the PID, when we ping to check if its alive, it would override PID with something incorrect
			// pid is garaunteed to always exist for APP_KEEPALIVE_PID_PATROL
			PID: self.o.GetPID(),
			Stopped: &Timestamp{
				Time:            now,
				TimestampFormat: self.patrol.config.Timestamp,
			},
			Disabled: self.o.IsDisabled(),
			Restart:  self.o.IsRestart(),
			// we want to know if we CONSUMED run_once, not if run_once is currently true!!!
			RunOnce:  self.o.IsRunOnceConsumed(),
			Shutdown: self.patrol.shutdown,
			// exit code is only garaunteed to exist for APP_KEEPALIVE_PID_PATROL
			ExitCode: self.o.GetExitCode(),
			KeyValue: self.o.GetKeyValue(),
		}
		if !self.o.GetStarted().IsZero() {
			h.Started = &Timestamp{
				Time:            self.o.GetStarted(),
				TimestampFormat: self.patrol.config.Timestamp,
			}
		}
		if self.o.GetLastSeen().IsZero() {
			if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
				// if our app was running lastseen should exist
				if !self.o.GetStarted().IsZero() {
					// we should set lastseen to now
					// we're responsible for this service to always be running
					h.LastSeen = &Timestamp{
						Time:            now,
						TimestampFormat: self.patrol.config.Timestamp,
					}
				}
			}
		} else {
			h.LastSeen = &Timestamp{
				Time:            self.o.GetLastSeen(),
				TimestampFormat: self.patrol.config.Timestamp,
			}
		}
		self.history = append(self.history, h)
		// reset values
		self.o.SetStarted(time.Time{})
		self.o.SetLastSeen(time.Time{})
		// do not unset started_log!!!
		// if our App forks, our App might still be using this log
		if self.o.IsRunOnceConsumed() {
			// we have to disable our app!
			self.toggle(API_TOGGLE_STATE_DISABLE)
		}
		self.o.SetPID(0)
		self.o.SetExitCode(0)
		if self.config.KeyValueClear {
			// clear keyvalues
			self.o.ReplaceKeyValue(nil)
		}
		// we're not going to use a goroutine here
		// we're assumed to be in a lock
		// we're going to unlock and then relock so that we can call our trigger
		if self.config.TriggerClosed != nil {
			self.o.Unlock()
			self.config.TriggerClosed(self, h)
			self.o.Lock()
		}
	}
}
func (self *App) startApp() error {
	now := time.Now()
	// consume restart
	self.o.SetRestart(false)
	// consume runonce
	if self.o.IsRunOnce() {
		self.o.SetRunOnceConsumed(true)
	}
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
		// when we build a command to execute, go will populate our args with that command
		// if we don't append our args to our current args we won't be able to see the process that we're executing
		// when we do `ps aux | grep -i binary` for example, 'binary' will be missing but all of the extra args will be present!
		// another reason we don't want to override this is that our first arg will be an exact path to our App/Binary
		// we can use this exact path to verify our PID in the future when using APP_KEEPALIVE_PID_APP
		// however, if we were to fork I'm unsure if first arg would remain the same, we could still verify this so long as our Binary value was contained in it!
		cmd.Args = append(cmd.Args, self.config.Args...)
	}
	if self.config.ExtraArgs != nil {
		if a := self.config.ExtraArgs(self); len(a) > 0 {
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
		if e := self.config.ExtraEnv(self); len(e) > 0 {
			cmd.Env = append(cmd.Env, e...)
		}
	}
	// patrol environment variables
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", APP_ENV_APP_ID, self.id))
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, self.config.KeepAlive))
	if self.config.PIDPath != "" {
		// pid path
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", APP_ENV_PID, filepath.Clean(self.config.WorkingDirectory+"/"+self.config.PIDPath)))
	}
	if len(self.patrol.config.ListenHTTP) > 0 {
		// http listeners
		bs, _ := json.Marshal(self.patrol.config.ListenHTTP)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", APP_ENV_LISTEN_HTTP, bs))
	}
	if len(self.patrol.config.ListenUDP) > 0 {
		// udp listeners
		bs, _ := json.Marshal(self.patrol.config.ListenUDP)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", APP_ENV_LISTEN_UDP, bs))
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
		ld := logDir(
			now,
			self.config.WorkingDirectory,
			self.config.LogDirectory,
		)
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
		ld := logDir(
			now,
			self.config.WorkingDirectory,
			self.config.LogDirectory,
		)
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
		if e := self.config.ExtraFiles(self); len(e) > 0 {
			cmd.ExtraFiles = e
		}
	}
	// we still have to set our WorkingDirectory
	cmd.Dir = self.config.WorkingDirectory
	// SysProcAttr holds optional operating system-specific attributes.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// so long as our KeepAlive method IS NOT APP_KEEPALIVE_PID_PATROL our children will have their own process group IDs
		// if we're using APP_KEEPALIVE_PID_PATROL we want to also receive any signals sent to our patrol PGID
		//
		// in any scenario, we will always make a best attempt to signal our children on shutdown or if our parent exits
		// APP_KEEPALIVE_PID_PATROL is the exception to this, we can't garauntee a children process will be signalled or even that that child will handle our signal!
		// if this is truly important to you, you should use a different KeepAlive method!
		// ideally in the future we may want to add support for something similar to `killall APP` on initial patrol start
		// we may also want to look into lock files as well as writing our PID to file when using APP_KEEPALIVE_PID_PATROL
		//
		// as of now, the only way we can really overcome this is to:
		// 1. share the same process group
		// 2. patrol signals children on close
		// 3. use processes that won't break if a parallel process is run
		//
		Setpgid: self.config.KeepAlive != APP_KEEPALIVE_PID_PATROL,
		// PGID should never be set
		Pgid: 0,
		// Signal that the process will get when its parent dies (Linux only)
		//
		// this signal should not be relied upon for shutting down your App, it is a courtesy
		//
		// The SIGTERM signal is a generic signal used to cause program termination.
		// This signal can be blocked, handled, and ignored. It is the normal way to politely ask a program to terminate
		// we don't have to close our process, but we should be aware that we're not being monitored
		// some processes may notice they receive 2 SIGTERMS, I'm not sure why it's doing this, just ignore additional signals
		Pdeathsig: syscall.SIGTERM,
	}
	if self.patrol.config.unittesting {
		// WE'RE UNITTESTING!!!
		// we DO NOT want to send a signal on death!!
		// we need predictable behavior since we will kill both parent and child process to test patrol
		cmd.SysProcAttr.Pdeathsig = 0
	}
	// start will start our process but will not wait for execute to finish running
	if err := cmd.Start(); err != nil {
		// failed to start
		return err
	}
	// started!
	self.o.SetStarted(now)
	self.o.SetStartedLog(now)
	if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// we're going to copy our PID from our process
		// any other keep alive method we're just going to ignore the process PID and assume it's wrong
		self.o.SetPID(uint32(cmd.Process.Pid))
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
		self.o.Lock()
		// set exit code
		self.o.SetExitCode(exit_code)
		// close app
		self.close()
		self.o.Unlock()
	}()
	return nil
}
func (self *App) isAppRunning() error {
	// check
	if self.config.KeepAlive == APP_KEEPALIVE_HTTP ||
		self.config.KeepAlive == APP_KEEPALIVE_UDP {
		// check if we've been pinged recently
		// if lastseen + ping timeout is NOT after now we know that we've timedout
		if self.o.GetLastSeen().IsZero() {
			// use started timestamp
			if time.Now().After(self.o.GetStarted().Add(time.Duration(self.patrol.config.PingTimeout) * time.Second)) {
				// expired
				// close app
				self.close()
				return ERR_APP_PING_EXPIRED
			}
		} else {
			// use lastseen
			if time.Now().After(self.o.GetLastSeen().Add(time.Duration(self.patrol.config.PingTimeout) * time.Second)) {
				// expired
				// close app
				self.close()
				return ERR_APP_PING_EXPIRED
			}
		}
		// running!
		return nil
	} else if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// check our internal state
		if self.o.GetStarted().IsZero() {
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
	// compare our PID
	if self.o.GetPID() > 0 {
		// App PID exists
		if pid != self.o.GetPID() {
			// App PID does not match
			// close previous App
			self.close()
			// set PID
			self.o.SetPID(pid)
			// this is a new App
			self.o.SetStarted(now)
			// we need to call our started trigger
			if self.config.TriggerStarted != nil {
				self.o.Unlock()
				self.config.TriggerStarted(self)
				self.o.Lock()
			}
		} else {
			// PID matches
			// app was previously started
			self.o.SetLastSeen(now)
		}
	} else {
		// App PID does not exist
		// set PID
		self.o.SetPID(pid)
		if self.o.GetStarted().IsZero() {
			// this is a new App
			self.o.SetStarted(now)
			// we need to call our started trigger
			if self.config.TriggerStarted != nil {
				self.o.Unlock()
				self.config.TriggerStarted(self)
				self.o.Lock()
			}
		} else {
			// app was previously started
			self.o.SetLastSeen(now)
		}
	}
	return nil
}
func (self *App) signalStop() {
	// we're going to signal our App if our App is disabled OR if we're shutting down Patrol
	// we can only do this if we have a PID, we don't care what keepalive method we use so long as a PID exists
	// we're going to discard any errors
	if self.o.GetPID() > 0 {
		if process, err := os.FindProcess(int(self.o.GetPID())); err == nil {
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
	// THIS FUNCTION SHOULD NOT SET A PID OR MODIFY LASTSEEN OR STARTED!!!
	// we should also NOT call triggers!!!
	// only once our PID is returned should we determine if we're a new or old process
	//
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
	return uint32(pid), nil
}
func (self *App) GetPID() uint32 {
	// this may not be the latest PID but it's the latest PID we're aware of
	self.o.RLock()
	defer self.o.RUnlock()
	return self.o.GetPID()
}
func (self *App) logDir() string {
	return logDir(
		self.o.GetStartedLog(),
		self.config.WorkingDirectory,
		self.config.LogDirectory,
	)
}
func logDir(
	started_log time.Time,
	wd string,
	ld string,
) string {
	y, m, d := started_log.Date()
	return filepath.Clean(
		fmt.Sprintf(
			"%s/%s/%d/%s/%d",
			wd,
			ld,
			y,
			strings.ToLower(m.String()),
			d,
		),
	)
}
func (self *App) GetStdoutLog() string {
	self.o.RLock()
	defer self.o.RUnlock()
	if self.config.Stdout != nil {
		// we don't know where the log is located
		return ""
	}
	if self.o.GetStartedLog().IsZero() {
		// we never started this app
		return ""
	}
	// we know where our logs are
	// THIS IS OUR LAST KNOWN LOCATION
	// IF WE FORK OUR PROCESS, OUR PROCESS MAY NOT PASS STDOUT/STDERR - THEN THIS IS USELESS!!
	// in our GUI we will offer a fallback to list all of our logs ideally
	return fmt.Sprintf("%s/%d.stdout.log", self.logDir(), self.o.GetStartedLog().UnixNano())
}
func (self *App) GetStderrLog() string {
	self.o.RLock()
	defer self.o.RUnlock()
	if self.config.Stderr != nil {
		// we don't know where the log is located
		return ""
	}
	if self.o.GetStartedLog().IsZero() {
		// we never started this app
		return ""
	}
	// we know where our logs are
	// THIS IS OUR LAST KNOWN LOCATION
	// IF WE FORK OUR PROCESS, OUR PROCESS MAY NOT PASS STDOUT/STDERR - THEN THIS IS USELESS!!
	// in our GUI we will offer a fallback to list all of our logs ideally
	return fmt.Sprintf("%s/%d.stderr.log", self.logDir(), self.o.GetStartedLog().UnixNano())
}
