package patrol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sabey.co/unittest"
	"syscall"
	"testing"
	"time"
)

func TestServerExecHTTP(t *testing.T) {
	log.Println("TestServerExecHTTP")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestServerExecHTTP")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)
	wd += "/unittest/"

	dir := fmt.Sprintf("%stestapp/", wd)
	x := fmt.Sprintf("%stestapp", dir)
	log.Printf("executing %s", x)
	cmd_app := exec.Command(x)
	cmd_app.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=http-testapp", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_HTTP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_app.Dir = dir
	cmd_app.Stdout = os.Stdout
	cmd_app.Stderr = os.Stderr
	cmd_app.Start()
	defer cmd_app.Process.Kill()
	go func() {
		cmd_app.Wait()
	}()

	<-time.After(time.Second * 2)

	dir = fmt.Sprintf("%stestserver/", wd)
	x = fmt.Sprintf("%stestserver", dir)
	log.Printf("executing %s", x)
	cmd_server := exec.Command(x, "-config", "config-http.json")
	cmd_server.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=testserver", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_HTTP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_server.Dir = dir
	cmd_server.Stdout = os.Stdout
	cmd_server.Stderr = os.Stderr
	cmd_server.Start()
	defer cmd_server.Process.Kill()
	go func() {
		cmd_server.Wait()
	}()

	// we have to wait at least PingTimeout
	<-time.After(time.Second * 3)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err := httpGET("http-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result := &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("closing testapp")
	cmd_app.Process.Kill()

	// we need to wait longer than PingTimeout!
	<-time.After(time.Second * 5)

	// check that we're closed
	log.Println("checking that we're closed")
	// hit API
	body, err = httpGET("http-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, len(result.History) > 0, true)
	unittest.Equals(t, result.History[0].Started.IsZero(), false)
	unittest.Equals(t, result.History[0].PID > 0, true)
	unittest.Equals(t, result.History[0].RunOnce, true)

	log.Println("enabling app")

	body, err = httpGET("http-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))

	// wait to be restarted
	log.Println("waiting for app to be restarted")

	// we have to wait at least PingTimeout + wait for a Ping (we pause 2 seconds)
	<-time.After(time.Second * 7)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err = httpGET("http-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("killing patrol")
	cmd_server.Process.Kill()

	log.Println("checking that we're running")
	<-time.After(time.Second * 3)

	// we can't PING our API anymore, we have to `kill -0 PID`
	process, err := os.FindProcess(int(result.PID))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	log.Println("still running! closing app")
	process.Signal(syscall.SIGKILL)
}
func TestServerExecUDP(t *testing.T) {
	log.Println("TestServerExecUDP")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestServerExecUDP")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)
	wd += "/unittest/"

	dir := fmt.Sprintf("%stestapp/", wd)
	x := fmt.Sprintf("%stestapp", dir)
	log.Printf("executing %s", x)
	cmd_app := exec.Command(x)
	cmd_app.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=udp-testapp", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_UDP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_app.Dir = dir
	cmd_app.Stdout = os.Stdout
	cmd_app.Stderr = os.Stderr
	cmd_app.Start()
	defer cmd_app.Process.Kill()
	go func() {
		cmd_app.Wait()
	}()

	<-time.After(time.Second * 2)

	dir = fmt.Sprintf("%stestserver/", wd)
	x = fmt.Sprintf("%stestserver", dir)
	log.Printf("executing %s", x)
	cmd_server := exec.Command(x, "-config", "config-udp.json")
	cmd_server.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=testserver", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_UDP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_server.Dir = dir
	cmd_server.Stdout = os.Stdout
	cmd_server.Stderr = os.Stderr
	cmd_server.Start()
	defer cmd_server.Process.Kill()
	go func() {
		cmd_server.Wait()
	}()

	// we have to wait at least PingTimeout
	<-time.After(time.Second * 3)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err := httpGET("udp-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result := &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("closing testapp")
	cmd_app.Process.Kill()

	// we need to wait longer than PingTimeout!
	<-time.After(time.Second * 5)

	// check that we're closed
	log.Println("checking that we're closed")
	// hit API
	body, err = httpGET("udp-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, len(result.History) > 0, true)
	unittest.Equals(t, result.History[0].Started.IsZero(), false)
	unittest.Equals(t, result.History[0].PID > 0, true)
	unittest.Equals(t, result.History[0].RunOnce, true)

	log.Println("enabling app")

	body, err = httpGET("udp-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))

	// wait to be restarted
	log.Println("waiting for app to be restarted")

	// we have to wait at least PingTimeout + wait for a Ping (we pause 2 seconds)
	<-time.After(time.Second * 7)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err = httpGET("udp-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("killing patrol")
	cmd_server.Process.Kill()

	log.Println("checking that we're running")
	<-time.After(time.Second * 3)

	// we can't PING our API anymore, we have to `kill -0 PID`
	process, err := os.FindProcess(int(result.PID))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	log.Println("still running! closing app")
	process.Signal(syscall.SIGKILL)
}
func TestServerExecPIDAPP(t *testing.T) {
	log.Println("TestServerExecPIDAPP")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestServerExecPIDAPP")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)
	wd += "/unittest/"

	dir := fmt.Sprintf("%stestapp/", wd)
	x := fmt.Sprintf("%stestapp", dir)
	log.Printf("executing %s", x)
	cmd_app := exec.Command(x)
	cmd_app.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=pid-app-testapp", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_PID_APP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_app.Dir = dir
	cmd_app.Stdout = os.Stdout
	cmd_app.Stderr = os.Stderr
	cmd_app.Start()
	defer cmd_app.Process.Kill()
	go func() {
		cmd_app.Wait()
	}()

	<-time.After(time.Second * 2)

	dir = fmt.Sprintf("%stestserver/", wd)
	x = fmt.Sprintf("%stestserver", dir)
	log.Printf("executing %s", x)
	cmd_server := exec.Command(x, "-config", "config-pid-app.json")
	cmd_server.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=testserver", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_PID_APP),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_server.Dir = dir
	cmd_server.Stdout = os.Stdout
	cmd_server.Stderr = os.Stderr
	cmd_server.Start()
	defer cmd_server.Process.Kill()
	go func() {
		cmd_server.Wait()
	}()

	// we have to wait at least PingTimeout
	<-time.After(time.Second * 3)

	// check that we're running
	log.Println("checking that we're running")
	// hit API - mark as run once as well
	body, err := httpGET("pid-app-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result := &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, false)

	log.Println("closing testapp")
	cmd_app.Process.Kill()

	// we need to wait longer than PingTimeout!
	<-time.After(time.Second * 5)

	// check that we're closed
	log.Println("checking that we're closed")
	// hit API
	body, err = httpGET("pid-app-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, len(result.History) > 0, true)
	unittest.Equals(t, result.History[0].Started.IsZero(), false)
	unittest.Equals(t, result.History[0].PID > 0, true)
	unittest.Equals(t, result.History[0].RunOnce, true)

	log.Println("enabling app")

	body, err = httpGET("pid-app-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))

	// wait to be restarted
	log.Println("waiting for app to be restarted")

	// we have to wait at least PingTimeout + wait for a Ping (we pause 2 seconds)
	<-time.After(time.Second * 7)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err = httpGET("pid-app-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("killing patrol")
	cmd_server.Process.Kill()

	log.Println("checking that we're running")
	<-time.After(time.Second * 3)

	// we can't PING our API anymore, we have to `kill -0 PID`
	process, err := os.FindProcess(int(result.PID))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	log.Println("still running! closing app")
	process.Signal(syscall.SIGKILL)
}
func TestServerExecPIDPatrol(t *testing.T) {
	log.Println("TestServerExecPIDPatrol")

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-time.After(time.Second * 45):
			log.Fatalln("failed to complete TestServerExecPIDPatrol")
		case <-done:
			return
		}
	}()

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)
	wd += "/unittest/"

	dir := fmt.Sprintf("%stestserver/", wd)
	x := fmt.Sprintf("%stestserver", dir)
	log.Printf("executing %s", x)
	cmd_server := exec.Command(x, "-config", "config-pid-patrol.json")
	cmd_server.Env = []string{
		fmt.Sprintf("%s=%s", PATROL_ENV_UNITTEST_KEY, PATROL_ENV_UNITTEST_VALUE),
		fmt.Sprintf("%s=testserver", APP_ENV_APP_ID),
		fmt.Sprintf("%s=%d", APP_ENV_KEEPALIVE, APP_KEEPALIVE_PID_PATROL),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_HTTP, LISTEN_HTTP_PORT_DEFAULT),
		fmt.Sprintf("%s=[\":%d\"]", APP_ENV_LISTEN_UDP, LISTEN_UDP_PORT_DEFAULT),
	}
	cmd_server.Dir = dir
	cmd_server.Stdout = os.Stdout
	cmd_server.Stderr = os.Stderr
	cmd_server.Start()
	defer cmd_server.Process.Kill()
	go func() {
		cmd_server.Wait()
	}()

	// we have to wait at least PingTimeout
	<-time.After(time.Second * 3)

	// check that we're running
	log.Println("checking that we're running")
	// hit API - mark as run once as well
	body, err := httpGET("pid-patrol-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result := &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, false)

	log.Println("closing testapp")
	process, err := os.FindProcess(int(result.PID))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.SIGKILL))

	// we need to wait longer than PingTimeout!
	<-time.After(time.Second * 5)

	// check that we're closed
	log.Println("checking that we're closed")
	// hit API
	body, err = httpGET("pid-patrol-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), true)
	unittest.Equals(t, result.Disabled, true)
	unittest.Equals(t, len(result.History) > 0, true)
	unittest.Equals(t, result.History[0].Started.IsZero(), false)
	unittest.Equals(t, result.History[0].PID > 0, true)
	unittest.Equals(t, result.History[0].RunOnce, true)

	log.Println("enabling app")

	body, err = httpGET("pid-patrol-testapp", fmt.Sprintf("toggle=%d", API_TOGGLE_STATE_ENABLE_RUNONCE_ENABLE))
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))

	// wait to be restarted
	log.Println("waiting for app to be restarted")

	// we have to wait at least PingTimeout + wait for a Ping (we pause 2 seconds)
	<-time.After(time.Second * 7)

	// check that we're running
	log.Println("checking that we're running")
	// hit API
	body, err = httpGET("pid-patrol-testapp", "")
	unittest.IsNil(t, err)
	log.Printf("body: \"%s\"\n", body)
	result = &API_Response{}
	unittest.IsNil(t, json.Unmarshal(body, result))
	// check response
	unittest.Equals(t, result.Started.IsZero(), false)
	unittest.Equals(t, result.PID > 0, true)
	unittest.Equals(t, result.RunOnce, true)

	log.Println("killing patrol")
	cmd_server.Process.Kill()

	log.Println("checking that we're running")
	<-time.After(time.Second * 2)

	// our process should still be running!!!
	// the reason it is still running is that we sent our signal to our patrol process and not patrol process group id
	// we have also disabled all forms of having patrol signal children to stop on shutdown
	// this is the downside of APP_KEEPALIVE_PID_PATROL, there is no way to tie a child to the life of a parent process?

	// we can't PING our API anymore, we have to `kill -0 PID`
	process, err = os.FindProcess(int(result.PID))
	unittest.IsNil(t, err)
	unittest.NotNil(t, process)
	unittest.IsNil(t, process.Signal(syscall.Signal(0)))

	log.Println("still running! closing app")
	unittest.IsNil(t, process.Signal(syscall.SIGKILL))
}

func httpGET(
	id string,
	extra string,
) (
	[]byte,
	error,
) {
	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/?id=%s&group=app&%s", LISTEN_HTTP_PORT_DEFAULT, id, extra))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		log.Fatalf("httpGET failed - StatusCode: %d != 200\n", response.StatusCode)
		return nil, nil
	}
	// read body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		response.Body.Close()
		return nil, err
	}
	response.Body.Close()
	return body, nil
}

// APP_KEEPALIVE_PID_PATROL, APP_KEEPALIVE_HTTP, APP_KEEPALIVE_UDP
//
// processes created by unittests:
//
// PPID   PID  PGID   SID TTY      TPGID STAT   UID   TIME COMMAND
// 4543  8092  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |   \_ go test -test.run=TestServerExecHTTP -race
// 8092  8268  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |       \_ /tmp/go-build905010232/b001/patrol.test -test.run=TestServerExecHTTP
// 8268  8274  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |           \_ /home/jackson/go/src/sabey.co/patrol/unittest/testapp/testapp
// 7388  8284  8283  7388 pts/14    8283 S+    1000   0:00  |       \_ grep --color=auto -i test
//
//
// processes created by patrol:
//
// PPID   PID  PGID   SID TTY      TPGID STAT   UID   TIME COMMAND
// 4543  8092  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |   \_ go test -test.run=TestServerExecHTTP -race
// 8092  8268  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |       \_ /tmp/go-build905010232/b001/patrol.test -test.run=TestServerExecHTTP
// 8268  8285  8092  4543 pts/0     8092 Sl+   1000   0:00  |   |           \_ /home/jackson/go/src/sabey.co/patrol/unittest/testserver/testserver -config config-http.json
// 8285  8313  8313  4543 pts/0     8092 Sl    1000   0:00  |   |               \_ /home/jackson/go/src/sabey.co/patrol/unittest/testapp/testapp
// 7388  8328  8327  7388 pts/14    8327 S+    1000   0:00  |       \_ grep --color=auto -i test
