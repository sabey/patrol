package cas

import (
	"log"
	"sabey.co/unittest"
	"sync"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	log.Println("TestService")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test service")
		case <-done:
			return
		}
	}()

	s := CreateService(false)
	cas := s.cas
	unittest.Equals(t, cas > 0, true)
	unittest.Equals(t, s.locked_read, false)
	unittest.Equals(t, s.locked_write, false)

	s.Lock()
	unittest.Equals(t, s.locked_read, true)
	unittest.Equals(t, s.locked_write, true)
	s.SetStarted(time.Time{})
	s.SetLastSeen(time.Time{})
	s.SetDisabled(false)
	s.SetRestart(false)
	s.SetRunOnce(false)
	s.SetRunOnceConsumed(false)
	unittest.Equals(t, s.incremented, false)
	unittest.Equals(t, s.cas, cas)
	s.Unlock()

	unittest.Equals(t, s.locked_read, false)
	unittest.Equals(t, s.locked_write, false)
	unittest.Equals(t, s.incremented, false)
	unittest.Equals(t, s.cas, cas)

	cas = s.cas
	s.Lock()
	s.SetStarted(time.Now())
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetStarted(time.Now())
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetLastSeen(time.Now())
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetLastSeen(time.Now())
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetDisabled(true)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetDisabled(false)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetRestart(true)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetRestart(false)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetRunOnce(true)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetRunOnce(false)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetRunOnceConsumed(true)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetRunOnceConsumed(false)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.SetKeyValue(nil)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.SetKeyValue(nil)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()

	cas = s.cas
	s.Lock()
	s.ReplaceKeyValue(nil)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.ReplaceKeyValue(nil)
	unittest.Equals(t, s.incremented, true)
	unittest.Equals(t, s.cas, cas+1)
	s.Unlock()
}
func TestServicePanic(t *testing.T) {
	log.Println("TestServicePanic")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test service panic")
		case <-done:
			return
		}
	}()

	var wg sync.WaitGroup
	c := 17
	wg.Add(c)

	i := 0
	var i_mu sync.Mutex
	s := CreateService(false)

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
		s.GetCAS()
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
		s.GetKeyValue()
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
		s.GetStarted()
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
		s.GetLastSeen()
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
		s.IsDisabled()
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
		s.IsRestart()
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
		s.IsRunOnce()
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
		s.IsRunOnceConsumed()
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
		s.Increment()
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
		s.SetKeyValue(nil)
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
		s.ReplaceKeyValue(nil)
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
		s.SetStarted(time.Time{})
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
		s.SetLastSeen(time.Time{})
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
		s.SetDisabled(false)
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
		s.SetRestart(false)
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
		s.SetRunOnce(false)
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
		s.SetRunOnceConsumed(false)
	}()

	// wait
	wg.Wait()
	unittest.Equals(t, i, c)
}
func TestServiceSetPanic(t *testing.T) {
	log.Println("TestServiceSetPanic")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 120):
			log.Fatalln("failed to complete test service set panic")
		case <-done:
			return
		}
	}()

	var wg sync.WaitGroup
	c := 9
	wg.Add(c)

	i := 0
	var i_mu sync.Mutex
	s := CreateService(false)
	s.RLock()
	defer s.RUnlock()
	unittest.Equals(t, s.locked_read, true)
	unittest.Equals(t, s.locked_write, false)

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
		s.Increment()
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
		s.SetKeyValue(nil)
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
		s.ReplaceKeyValue(nil)
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
		s.SetStarted(time.Time{})
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
		s.SetLastSeen(time.Time{})
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
		s.SetDisabled(false)
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
		s.SetRestart(false)
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
		s.SetRunOnce(false)
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
		s.SetRunOnceConsumed(false)
	}()

	// wait
	wg.Wait()
	unittest.Equals(t, i, c)
}
