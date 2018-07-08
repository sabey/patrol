package patrol

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
	history  []*History
	started  time.Time
	disabled bool // takes its initial value from config
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
func (self *App) GetHistory() []*History {
	// dereference
	history := make([]*History, 0, len(self.history))
	for _, h := range self.history {
		history = append(history, h.clone())
	}
	return history
}
func (self *App) saveHistory(
	shutdown bool,
) {
	if !self.started.IsZero() {
		// save history
		if len(self.history) >= self.patrol.config.History {
			self.history = self.history[1:]
		}
		h := &History{
			Started:  self.started,
			Stopped:  time.Now(),
			Shutdown: shutdown,
		}
		self.history = append(self.history, h)
		// unset previous started so we don't create duplicate histories
		self.started = time.Time{}
		// call trigger in a go routine so we don't deadlock
		if self.config.TriggerStopped != nil {
			go self.config.TriggerStopped(self.id, self, h)
		}
	}
}
func (self *App) startApp() error {
	// we are ASSUMING our app isn't started!!!
	// this function should only ever be called by tick()
	// we gotta lock and defer to set history
	self.mu.Lock()
	defer self.mu.Unlock()
	// save history
	self.saveHistory(false)
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
	// STD in/out/err
	if self.config.Stdin != nil {
		cmd.Stdin = self.config.Stdin
	}
	if self.config.Stdout != nil {
		cmd.Stdout = self.config.Stdout
	}
	if self.config.Stderr != nil {
		cmd.Stderr = self.config.Stderr
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
	self.is_running = true
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
		cmd.Wait()
		// currently this can't race because we ALWAYS check isAppRunning() before startApp() AND we only use tick() to start services
		// this logic should never change, so it's not something to worry about right now
		self.mu.Lock()
		self.is_running = false
		// save history
		self.saveHistory(false)
		self.mu.Unlock()
	}()
	return nil
}
func (self *App) isAppRunning() error {
	// if we determine a process is NOT running we will set history - we will NOT attempt to restart anything!
	// lock and defer incase we have to set history!
	self.mu.Lock()
	defer self.mu.Unlock()
	// check
	if self.config.KeepAlive == APP_KEEPALIVE_HTTP ||
		self.config.KeepAlive == APP_KEEPALIVE_UDP {
		// check if we've been pinged recently
		// if last ping + ping timeout is NOT after now we know that we've timedout
		if time.Now().After(self.ping.Add(APP_PING_EVERY)) {
			// expired
			// save history
			self.saveHistory(false)
			return ERR_APP_PING_EXPIRED
		}
		// still alive
		if self.started.IsZero() {
			// we need to set started since this is our first time seeing this app
			self.started = time.Now()
		}
		return nil
	} else if self.config.KeepAlive == APP_KEEPALIVE_PID_PATROL {
		// check our internal state
		if !self.is_running {
			// not running
			// we do NOT have to save history!!!
			// our teardown function after cmd.Wait() will save our history!
			return ERR_APP_KEEPALIVE_PATROL_NOTRUNNING
		}
		// running
		if self.started.IsZero() {
			// we need to set started since this is our first time seeing this app
			self.started = time.Now()
		}
		return nil
	}
	// we have to ping our PID to determine if we're running
	// this function is only used by APP_KEEPALIVE_PID_APP
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	pid, err := self.getPID()
	if err != nil {
		// failed to find PID
		// save history
		self.saveHistory(false)
		return err
	}
	// TODO: we should add PID verification here
	// either before or after we signal to kill, it's unsure how this will work
	cmd := exec.CommandContext(ctx, "kill", "-0", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		// NOT running!
		// save history
		self.saveHistory(false)
		return err
	}
	// running!
	if self.started.IsZero() {
		// we need to set started since this is our first time seeing this app
		self.started = time.Now()
	}
	return nil
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
	return uint32(pid), nil
}
func (self *App) GetPID() uint32 {
	// this may not be the latest PID but it's the latest PID we're aware of
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.pid
}
