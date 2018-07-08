package patrol

import (
	"fmt"
	"log"
	"os"
	"sabey.co/unittest"
	"testing"
)

func TestPatrolApp(t *testing.T) {
	log.Println("TestPatrolApp")

	wd, err := os.Getwd()
	unittest.IsNil(t, err)
	unittest.Equals(t, wd != "", true)

	config := &ConfigApp{
		KeepAlive: APP_KEEPALIVE_PID_PATROL,
	}
	unittest.Equals(t, config.Validate(), ERR_APP_NAME_EMPTY)

	config.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, config.Validate(), ERR_APP_NAME_MAXLENGTH)

	config.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	unittest.Equals(t, config.Validate(), ERR_APP_WORKINGDIRECTORY_EMPTY)

	config.WorkingDirectory = "directory"
	unittest.Equals(t, config.Validate(), ERR_APP_WORKINGDIRECTORY_RELATIVE)

	config.WorkingDirectory = "/directory/."
	unittest.Equals(t, config.Validate(), ERR_APP_WORKINGDIRECTORY_UNCLEAN)

	config.WorkingDirectory = "/directory"
	unittest.Equals(t, config.Validate(), ERR_APP_BINARY_EMPTY)

	// setting working directory to cwd
	config.WorkingDirectory = wd + "/unittest"

	config.Binary = "file/.."
	unittest.Equals(t, config.Validate(), ERR_APP_BINARY_UNCLEAN)

	config.Binary = "file"
	unittest.Equals(t, config.Validate(), ERR_APP_LOGDIRECTORY_EMPTY)

	config.LogDirectory = "log-directory/."
	unittest.Equals(t, config.Validate(), ERR_APP_LOGDIRECTORY_UNCLEAN)

	config.LogDirectory = "log-directory"
	unittest.Equals(t, config.Validate(), ERR_APP_PIDPATH_EMPTY)

	config.PIDPath = "pid/.."
	unittest.Equals(t, config.Validate(), ERR_APP_PIDPATH_UNCLEAN)

	// changing pid to app.pid
	config.PIDPath = "app.pid"
	unittest.IsNil(t, config.Validate())

	app := &App{
		config: config,
	}

	fmt.Println("app.getPID")

	pid, err := app.getPID()
	unittest.IsNil(t, err)
	unittest.Equals(t, pid, 1254)

	app.config.PIDPath = "bad.pid"
	unittest.IsNil(t, app.config.Validate())
	_, err = app.getPID()
	unittest.NotNil(t, err)
}
