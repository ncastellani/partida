package bootstrap

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncastellani/partida/utilfunc"
)

type Application struct {
	Logger  *log.Logger            // default application logger
	Path    string                 // path to the directory that this app is running at
	Version string                 // current application version
	Config  map[string]interface{} // application general config
	Vars    map[string]string      // the config-requested environment variables

	// all of the the interfaces below this comment must be defined by the
	// client application after calling the NewApplication function

	Backend Backend         // map to the client application interfaces
	Codes   map[string]Code // map of the available response codes

	// API settings
	APILogsWriter io.Writer
	APIMethods    map[string]APIResourceMethod
	APIRoutes     map[string]map[string]APIResource
	APIValidators map[string]APIParameterValidator

	// Queue settings
	QueueLogsWriter io.Writer
	QueueMethods    map[string]QueueMethod
}

// NewApplication
// recieve an logger and a file path to the config JSON
// file and determine the required information to initiate an API
// and queue handler applcation.
// will panic if failure.
func NewApplication(l *log.Logger, config string) (app Application) {

	// current directory full path
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("failed to determine the current execution path [err: %v]", err)
	}

	app.Path = pwd

	l.Printf("found the current path of this application [pwd: %v]", app.Path)

	// determine the current server version or generate a new one
	app.Version = os.Getenv("APP_VERSION")
	if app.Version == "" {
		app.Version = strings.ToUpper(utilfunc.RandomString(4))

		// try to fetch version from file
		b, _ := os.ReadFile(app.Path + filepath.Join("/", "version.txt"))
		if len(b) != 0 {
			app.Version = string(b)
		}

		os.Setenv("APP_VERSION", app.Version)
	}

	// parse the general config files
	err = utilfunc.ParseJSON(app.Path+config, &app.Config)
	if err != nil {
		l.Fatalf("failed to parse the config JSON [err: %v]", err)
	}

	log.Println("configuration file parsed and imported")

	return
}

// CheckForVariables
// take a list of environment variables names and
// check if they have a defined value on the environment.
// if the variable is available, it is set at the app.Vars
// map, if not, an fatal error will occur.
// will panic if failure.
func (app *Application) CheckForVariables(list []string) {
	for _, rv := range list {
		v := os.Getenv(rv)
		if v == "" {
			log.Fatalf("an required environment variable is not set [var: %v]", rv)
		}

		app.Vars[rv] = v
	}
}
