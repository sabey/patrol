package cas

import (
	"log"
	"sabey.co/unittest"
	"sync"
	"testing"
	"time"
)

func TestApp(t *testing.T) {
	log.Println("TestApp")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test app")
		case <-done:
			return
		}
	}()

	a := CreateApp(false)
	cas := a.cas
	unittest.Equals(t, cas > 0, true)
	unittest.Equals(t, a.locked_read, false)
	unittest.Equals(t, a.locked_write, false)

	a.Lock()
	unittest.Equals(t, a.locked_read, true)
	unittest.Equals(t, a.locked_write, true)
	a.SetStarted(time.Time{})
	a.SetStartedLog(time.Time{})
	a.SetLastSeen(time.Time{})
	a.SetDisabled(false)
	a.SetRestart(false)
	a.SetRunOnce(false)
	a.SetRunOnceConsumed(false)
	a.SetPID(0)
	a.SetExitCode(0)
	unittest.Equals(t, a.incremented, false)
	unittest.Equals(t, a.cas, cas)
	a.Unlock()

	unittest.Equals(t, a.locked_read, false)
	unittest.Equals(t, a.locked_write, false)
	unittest.Equals(t, a.incremented, false)
	unittest.Equals(t, a.cas, cas)

	cas = a.cas
	a.Lock()
	a.SetStarted(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetStarted(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetStartedLog(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetStartedLog(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetLastSeen(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetLastSeen(time.Now())
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetDisabled(true)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetDisabled(false)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetRestart(true)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetRestart(false)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetRunOnce(true)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetRunOnce(false)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetRunOnceConsumed(true)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetRunOnceConsumed(false)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetPID(1)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetPID(0)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetExitCode(1)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetExitCode(0)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.SetKeyValue(nil)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.SetKeyValue(nil)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()

	cas = a.cas
	a.Lock()
	a.ReplaceKeyValue(nil)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.ReplaceKeyValue(nil)
	unittest.Equals(t, a.incremented, true)
	unittest.Equals(t, a.cas, cas+1)
	a.Unlock()
}
func TestAppPanic(t *testing.T) {
	log.Println("TestAppPanic")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test app panic")
		case <-done:
			return
		}
	}()

	var wg sync.WaitGroup
	c := 23
	wg.Add(c)

	i := 0
	var i_mu sync.Mutex
	a := CreateApp(false)

	// getters
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetCAS()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetKeyValue()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetStarted()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetStartedLog()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetLastSeen()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.IsDisabled()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.IsRestart()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.IsRunOnce()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.IsRunOnceConsumed()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetPID()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.GetExitCode()
	}()

	// setters
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.Increment()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetKeyValue(nil)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.ReplaceKeyValue(nil)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetStarted(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetStartedLog(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetLastSeen(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetDisabled(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRestart(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRunOnce(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRunOnceConsumed(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetPID(0)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetExitCode(0)
	}()

	// wait
	wg.Wait()
	unittest.Equals(t, i, c)
}
func TestAppSetPanic(t *testing.T) {
	log.Println("TestAppSetPanic")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test app set panic")
		case <-done:
			return
		}
	}()

	var wg sync.WaitGroup
	c := 12
	wg.Add(c)

	i := 0
	var i_mu sync.Mutex
	a := CreateApp(false)
	a.RLock()
	defer a.RUnlock()
	unittest.Equals(t, a.locked_read, true)
	unittest.Equals(t, a.locked_write, false)

	// setters
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.Increment()
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetKeyValue(nil)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.ReplaceKeyValue(nil)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetStarted(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetStartedLog(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetLastSeen(time.Time{})
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetDisabled(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRestart(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRunOnce(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetRunOnceConsumed(false)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetPID(0)
	}()
	go func() {
		defer func() {
			defer wg.Done()
			if r := recover(); r != nil {
				i_mu.Lock()
				i++
				i_mu.Unlock()
			}
		}()
		a.SetExitCode(0)
	}()

	// wait
	wg.Wait()
	unittest.Equals(t, i, c)
}
