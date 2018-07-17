package cas

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type App struct {
	// unsafe
	// our original App struct inside of patrol included everything here but History
	// we have NO INTEREST in ever including history here, we don't want to completely restructure our project
	// when we append History we WILL call Increment() just incase
	keyvalue    map[string]interface{}
	started     time.Time
	started_log time.Time
	lastseen    time.Time
	// disabled takes its initial value from our config
	disabled bool
	// restart will ALWAYS cause our app to become enabled!
	restart bool
	// run once has to be consumed
	// once our app is running we will instantly consume runonce, once stopped we will become disabled
	// if our app is stopped, we will start and consume runonce
	run_once          bool
	run_once_consumed bool
	// pid is set by all but only used by APP_KEEPALIVE_PID_APP to verify a process is running
	// last PID we've found and verified
	// the maximum PID value on a 32 bit system is 32767
	// the maximum PID value on a 64 bit system is 2^22
	// systems by default will default to 32767, but we should support up to uint32
	pid uint32
	// we're going to save our exit code for history
	// this is only supported by APP_KEEPALIVE_PID_PATROL
	exit_code uint8
	// CAS will be init with a random number
	// we never want to end up in a situation where we keep restarting patrol and we init with a CAS of 1
	// we could enter a scenario where we've read values, patrol restarts, and then we set with that CAS and it succeeds
	// these aren't the same states at all, random numbers can help us avoid this situation
	//
	// CAS MUST BE UPDATED IF WE EVER MODIFY ANY ATTRIBUTE!!!
	// IF WE EVER UNLOCK, CALL A TRIGGER, RELOCK, AND MODIFY AN ATTRIBUTED, WE HAVE TO INCREMENT CAS!!!
	// ideally our entry points for modification should only ever update our CAS once
	// however, it's better to update our CAS twice if we're unsure rather than miss a CAS
	// CAS is currently only used through API, so see `patrol/api.go` for more info
	//
	// to solve our issues with missing a CAS we've created external objects to manage our internal values and CAS
	// we have one edge case for when we WILL NOT update CAS!!!:
	// if we update a value but that value is the EXACT same as the previous value, we have no reason to modify our CAS
	// this may change in the future but currently I see it as a NOOP
	//
	// moving our CAS objects externally from Patrol should help solve all of our issues of being able to trust our CAS
	cas          uint64
	locked_read  bool
	locked_write bool
	incremented  bool
	mu_internal  sync.RWMutex
	mu_external  sync.RWMutex
}

func CreateApp(
	disabled bool,
) *App {
	return &App{
		keyvalue: make(map[string]interface{}),
		disabled: disabled,
		cas:      uint64(rand.Intn(cas_init_max_range)) + 1,
	}
}

func (self *App) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *App) Lock() {
	self.mu_external.Lock()
	self.mu_internal.Lock()
	self.locked_read = true
	self.locked_write = true
	self.incremented = false
	self.mu_internal.Unlock()
}
func (self *App) Unlock() {
	self.mu_external.Unlock()
	self.mu_internal.Lock()
	self.locked_read = false
	self.locked_write = false
	self.incremented = false
	self.mu_internal.Unlock()
}
func (self *App) RLock() {
	self.mu_external.RLock()
	self.mu_internal.Lock()
	self.locked_read = true
	self.mu_internal.Unlock()
}
func (self *App) RUnlock() {
	self.mu_external.RUnlock()
	self.mu_internal.Lock()
	self.locked_read = false
	self.mu_internal.Unlock()
}
func (self *App) increment() {
	if !self.incremented {
		self.incremented = true
		self.cas++
	}
}
func (self *App) Increment() {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.Increment(): not locked_write!")
	}
	// Increment should be called when we update a value in Patrol that relies on our CAS value here!
	// our use case for this is History, we have no interest in restructuring Patrol to include History here
	self.increment()
}
func (self *App) GetCAS() uint64 {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetCAS(): not locked_read!")
	}
	return self.cas
}
func (self *App) GetKeyValue() map[string]interface{} {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetKeyValue(): not locked_read!")
	}
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
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetKeyValue(): not locked_write!")
	}
	// we're not going to use compare, we will ALWAYS increment here!!
	// if we don't want to increment, use len before calling this function
	self.increment()
	// dereference
	for k, v := range kv {
		self.keyvalue[k] = v
	}
}
func (self *App) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.ReplaceKeyValue(): not locked_write!")
	}
	// we're not going to use compare, we will ALWAYS increment here!!
	// if we don't want to increment, use len before calling this function
	self.increment()
	self.keyvalue = make(map[string]interface{})
	// dereference
	for k, v := range kv {
		self.keyvalue[k] = v
	}
}
func (self *App) GetStarted() time.Time {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetStarted(): not locked_read!")
	}
	return self.started
}
func (self *App) SetStarted(
	started time.Time,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetStarted(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.started.Equal(started) {
		// NOOP
		return
	}
	self.increment()
	self.started = started
}
func (self *App) GetStartedLog() time.Time {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetStartedLog(): not locked_read!")
	}
	return self.started_log
}
func (self *App) SetStartedLog(
	started_log time.Time,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetStartedLog(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.started_log.Equal(started_log) {
		// NOOP
		return
	}
	self.increment()
	self.started_log = started_log
}
func (self *App) GetLastSeen() time.Time {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetLastSeen(): not locked_read!")
	}
	return self.lastseen
}
func (self *App) SetLastSeen(
	lastseen time.Time,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetLastSeen(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.lastseen.Equal(lastseen) {
		// NOOP
		return
	}
	self.increment()
	self.lastseen = lastseen
}
func (self *App) IsDisabled() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.IsDisabled(): not locked_read!")
	}
	return self.disabled
}
func (self *App) SetDisabled(
	disabled bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetDisabled(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.disabled == disabled {
		// NOOP
		return
	}
	self.increment()
	self.disabled = disabled
}
func (self *App) IsRestart() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.IsRestart(): not locked_read!")
	}
	return self.restart
}
func (self *App) SetRestart(
	restart bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetRestart(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.restart == restart {
		// NOOP
		return
	}
	self.increment()
	self.restart = restart
}
func (self *App) IsRunOnce() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.IsRunOnce(): not locked_read!")
	}
	return self.run_once
}
func (self *App) SetRunOnce(
	run_once bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetRunOnce(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.run_once == run_once {
		// NOOP
		return
	}
	self.increment()
	self.run_once = run_once
}
func (self *App) IsRunOnceConsumed() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.IsRunOnceConsumed(): not locked_read!")
	}
	return self.run_once_consumed
}
func (self *App) SetRunOnceConsumed(
	run_once_consumed bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetRunOnceConsumed(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.run_once_consumed == run_once_consumed {
		// NOOP
		return
	}
	self.increment()
	self.run_once_consumed = run_once_consumed
}
func (self *App) GetPID() uint32 {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetPID(): not locked_read!")
	}
	return self.pid
}
func (self *App) SetPID(
	pid uint32,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetPID(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.pid == pid {
		// NOOP
		return
	}
	self.increment()
	self.pid = pid
}
func (self *App) GetExitCode() uint8 {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.App.GetExitCode(): not locked_read!")
	}
	return self.exit_code
}
func (self *App) SetExitCode(
	exit_code uint8,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.App.SetExitCode(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.exit_code == exit_code {
		// NOOP
		return
	}
	self.increment()
	self.exit_code = exit_code
}
