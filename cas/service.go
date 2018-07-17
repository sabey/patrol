package cas

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type Service struct {
	// unsafe
	// our original Service struct inside of patrol included everything here but History
	// we have NO INTEREST in ever including history here, we don't want to completely restructure our project
	// when we append History we WILL call Increment() just incase
	keyvalue map[string]interface{}
	started  time.Time
	lastseen time.Time
	// disabled takes its initial value from our config
	disabled bool
	// restart will ALWAYS cause our service to become enabled!
	restart bool
	// run once has to be consumed
	// once our service is running we will instantly consume runonce, once stopped we will become disabled
	// if our service is stopped, we will start and consume runonce
	run_once          bool
	run_once_consumed bool
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

func CreateService(
	disabled bool,
) *Service {
	return &Service{
		keyvalue: make(map[string]interface{}),
		disabled: disabled,
		cas:      uint64(rand.Intn(cas_init_max_range)) + 1,
	}
}

func (self *Service) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Service) Lock() {
	self.mu_external.Lock()
	self.mu_internal.Lock()
	self.locked_read = true
	self.locked_write = true
	self.incremented = false
	self.mu_internal.Unlock()
}
func (self *Service) Unlock() {
	self.mu_external.Unlock()
	self.mu_internal.Lock()
	self.locked_read = false
	self.locked_write = false
	self.incremented = false
	self.mu_internal.Unlock()
}
func (self *Service) RLock() {
	self.mu_external.RLock()
	self.mu_internal.Lock()
	self.locked_read = true
	self.mu_internal.Unlock()
}
func (self *Service) RUnlock() {
	self.mu_external.RUnlock()
	self.mu_internal.Lock()
	self.locked_read = false
	self.mu_internal.Unlock()
}
func (self *Service) increment() {
	if !self.incremented {
		self.incremented = true
		self.cas++
	}
}
func (self *Service) Increment() {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.Increment(): not locked_write!")
	}
	// Increment should be called when we update a value in Patrol that relies on our CAS value here!
	// our use case for this is History, we have no interest in restructuring Patrol to include History here
	self.increment()
}
func (self *Service) GetCAS() uint64 {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.GetCAS(): not locked_read!")
	}
	return self.cas
}
func (self *Service) GetKeyValue() map[string]interface{} {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.GetKeyValue(): not locked_read!")
	}
	// dereference
	kv := make(map[string]interface{})
	for k, v := range self.keyvalue {
		kv[k] = v
	}
	return kv
}
func (self *Service) SetKeyValue(
	kv map[string]interface{},
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetKeyValue(): not locked_write!")
	}
	// we're not going to use compare, we will ALWAYS increment here!!
	// if we don't want to increment, use len before calling this function
	self.increment()
	// dereference
	for k, v := range kv {
		self.keyvalue[k] = v
	}
}
func (self *Service) ReplaceKeyValue(
	kv map[string]interface{},
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.ReplaceKeyValue(): not locked_write!")
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
func (self *Service) GetStarted() time.Time {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.GetStarted(): not locked_read!")
	}
	return self.started
}
func (self *Service) SetStarted(
	started time.Time,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetStarted(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.started.Equal(started) {
		// NOOP
		return
	}
	self.increment()
	self.started = started
}
func (self *Service) GetLastSeen() time.Time {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.GetLastSeen(): not locked_read!")
	}
	return self.lastseen
}
func (self *Service) SetLastSeen(
	lastseen time.Time,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetLastSeen(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.lastseen.Equal(lastseen) {
		// NOOP
		return
	}
	self.increment()
	self.lastseen = lastseen
}
func (self *Service) IsDisabled() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.IsDisabled(): not locked_read!")
	}
	return self.disabled
}
func (self *Service) SetDisabled(
	disabled bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetDisabled(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.disabled == disabled {
		// NOOP
		return
	}
	self.increment()
	self.disabled = disabled
}
func (self *Service) IsRestart() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.IsRestart(): not locked_read!")
	}
	return self.restart
}
func (self *Service) SetRestart(
	restart bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetRestart(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.restart == restart {
		// NOOP
		return
	}
	self.increment()
	self.restart = restart
}
func (self *Service) IsRunOnce() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.IsRunOnce(): not locked_read!")
	}
	return self.run_once
}
func (self *Service) SetRunOnce(
	run_once bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetRunOnce(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.run_once == run_once {
		// NOOP
		return
	}
	self.increment()
	self.run_once = run_once
}
func (self *Service) IsRunOnceConsumed() bool {
	self.mu_internal.RLock()
	defer self.mu_internal.RUnlock()
	if !self.locked_read {
		log.Panicln("./patrol/cas.Service.IsRunOnceConsumed(): not locked_read!")
	}
	return self.run_once_consumed
}
func (self *Service) SetRunOnceConsumed(
	run_once_consumed bool,
) {
	self.mu_internal.Lock()
	defer self.mu_internal.Unlock()
	if !self.locked_write {
		log.Panicln("./patrol/cas.Service.SetRunOnceConsumed(): not locked_write!")
	}
	// we're only going to increment if our values are different
	if self.run_once_consumed == run_once_consumed {
		// NOOP
		return
	}
	self.increment()
	self.run_once_consumed = run_once_consumed
}
