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
	APIRoutes     map[string]map[string]APIResource // (done by LoadJSONFiles) path to HTTP method to function method map
	APILogsWriter io.Writer
	APIMethods    map[string]APIResourceMethod
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
	app.Logger = l

	// current directory full path
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		app.Logger.Fatalf("failed to determine the current execution path [err: %v]", err)
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
		app.Logger.Fatalf("failed to parse the config JSON [err: %v]", err)
	}

	app.Logger.Println("configuration file parsed and imported")

	return
}

// CheckForVariables
// take a list of environment variables names and
// check if they have a defined value on the environment.
// if the variable is available, it is set at the app.Vars
// map, if not, an fatal error will occur.
// will panic if failure.
func (app *Application) CheckForVariables(list []string) {
	app.Vars = make(map[string]string)

	for _, rv := range list {
		v := os.Getenv(rv)
		if v == "" {
			app.Logger.Fatalf("an required environment variable is not set [var: %v]", rv)
		}

		app.Logger.Printf("found required environment variable [var: %v]", rv)
		app.Vars[rv] = v
	}
}

// LoadJSONFiles
// take the relative file path of the API codes and routes
// JSON file, import those contents and set them on the application.
// the imported codes are merged with the default ones.
// will panic if failure.
func (app *Application) LoadJSONFiles(codes, routes string) {
	app.Logger.Println("loading the JSON files with the codes and routes...")

	// import the API codes
	var parsedCodes map[string]Code

	err := utilfunc.ParseJSON(app.Path+codes, &parsedCodes)
	if err != nil {
		app.Logger.Fatalf("failed to import codes JSON file [err: %v]", err)
	}

	// merge the default codes with the imported ones
	for k, v := range DefaultCodes {
		parsedCodes[k] = v
	}

	// set the parsed codes in the application
	app.Codes = parsedCodes

	// load the API routes to the application
	err = utilfunc.ParseJSON(app.Path+routes, &app.APIRoutes)
	if err != nil {
		app.Logger.Fatalf("failed to import routes JSON file [err: %v]", err)
	}

}
