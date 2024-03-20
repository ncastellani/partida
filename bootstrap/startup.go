package bootstrap

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncastellani/partida/utilfunc"
)

type Application struct {
	Path    string // directory this app is running at
	Version string // current application version
	Config  map[string]interface{}

	ExecKind    string // execution type
	ExecHandler string // execution handler type
	Region      string // current region this server is running
	HTTPPort    string // server port to expose the HTTP handler
}

// !! will panic if failure.
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
	err = utilfunc.ParseJSON(app.Path+config, app.Config)
	if err != nil {
		l.Fatalf("failed to parse the config JSON [err: %v]", err)
	}

	log.Println("configuration file parsed and imported")

	return
}
