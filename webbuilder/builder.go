package webbuilder

import "io"

type BuildConfig struct {

	// IO writers to output the build proccess logs
	IODefault io.Writer
	IODebug   io.Writer

	// general web application data
	DistributionPath string                 // folder path to the compiled result folder. ex: ".dist"
	RenderParameters map[string]interface{} // extra parameters that will be passed into the templating engine
	Version          string                 // your current web application version. will be passed to the templating engine
	BaseView         string                 // the file that contains your base layout that all others views uses. ex: "_base.html"

	// web application source files folders
	Folder struct { // the local path into the nominated folder. ex: "assets/"
		Assets    string
		Languages string
		Views     string
		Static    string
	}

	// build generated structures
	languages []Language

	//
}
