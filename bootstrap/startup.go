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
	BuildID string                 // number or "version" of the current compilation ID (like GitHub Actions BuildID)
	Version string                 // current application version (like 1.0, 2.0.5 ...)
	Config  map[string]interface{} // application general config
	Vars    map[string]string      // the config-requested environment variables
	Codes   map[string]Code        // map of the available response codes
	Backend Backend                // map to the client application interfaces

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
// recieve an logger and the paths to the JSON config files. determine the required
// information to initiate an API and queue handler applcation. will panic if failure.
func NewApplication(l *log.Logger, config, codes, routes string) (app Application) {
	var err error

	// check if a logger was passed. if not, use the default one
	if l == nil {
		app.Logger = log.Default()
	} else {
		app.Logger = l
	}

	// current directory full path
	app.Path, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		app.Logger.Fatalf("failed to determine the current execution path [err: %v]", err)
	}

	app.Logger.Printf("found the current path of this application [pwd: %v]", app.Path)

	// determine the current build ID from env var or generate a new one
	app.BuildID = os.Getenv("APP_BUILD_ID")
	if app.BuildID == "" {
		app.BuildID = strings.ToUpper(utilfunc.RandomString(4))
		os.Setenv("APP_BUILD_ID", app.BuildID)
	}

	// determine the current server version from env var or file or use "0.1"
	app.Version = os.Getenv("APP_VERSION")
	if app.Version == "" {
		app.Version = "0.1"

		// try to fetch version from file
		b, _ := os.ReadFile(app.Path + filepath.Join("/", "version.txt"))
		if len(b) != 0 {
			app.Version = string(b)
		}

		os.Setenv("APP_VERSION", app.Version)
	}

	app.Logger.Printf("determined the current version and build ID [version: %v] [buildID: %v]", app.Version, app.BuildID)

	// parse the general config files
	err = utilfunc.ParseJSON(app.Path+config, &app.Config)
	if err != nil {
		app.Logger.Fatalf("failed to parse the config JSON [err: %v]", err)
	}

	app.Logger.Println("configuration file parsed and imported")

	// import the API codes
	var parsedCodes map[string]Code

	err = utilfunc.ParseJSON(app.Path+codes, &parsedCodes)
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

	app.Logger.Printf("loaded the JSON files with the application codes and API routes [availableCodes: %v] [availableAPIRoutes: %v]", len(parsedCodes), len(app.APIRoutes))

	return
}

// CheckForVariables
// take a list of environment variables names and check if they have a defined value.
// if the variable is available, it is set at the app.Vars map, if not, an fatal error will occur.
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
