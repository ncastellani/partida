package webbuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/ncastellani/partida/utilfunc"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
)

// runJob
// prepare the files and call the operators to run a compilation job.
// will panic if failure.
func (c *BuildConfig) runJob(job, folder, rebaseTo string, includeSuffix bool, suffixes []string, op func(l *log.Logger, f utilfunc.File, newPath string)) {

	// generate a job logger
	l := log.New(c.IODebug, fmt.Sprintf("[job: %v] ", job), log.LstdFlags|log.Lmsgprefix)

	l.Println("starting compile procedures...")

	// list files from the passed path
	files, err := utilfunc.ListFolderFiles(folder)
	if err != nil {
		l.Panicf("failed to list folder files [err: %v]", err)
		return
	}

	l.Printf("fetched the file list of the passed folder [count: %v]", len(files))

	// handle each file
	for _, f := range files {

		// skip folders
		if f.IsDir {
			continue
		}

		if len(suffixes) != 0 {
			skip := false
			for _, sf := range suffixes {
				if includeSuffix {
					if !strings.HasSuffix(f.Path, sf) {
						skip = true
					} else {
						skip = false
						break
					}
				} else {
					if strings.HasSuffix(f.Path, sf) {
						skip = true
					} else {
						skip = false
						break
					}
				}
			}

			if skip {
				continue
			}
		}

		// rebase its path and generate a logger for this file
		newPath := strings.Replace(filepath.ToSlash(f.Path), folder, c.DistributionPath+rebaseTo, 1)

		fl := log.New(c.IODebug, fmt.Sprintf("%v(%v) ", l.Prefix(), f.Path), l.Flags())
		fl.Printf("rebased file path [rebased: %v]", newPath)

		// call the operator function
		op(fl, f, newPath)

	}

}

// ImportLanguages
// scan the passed languages folder for JSON files
// and import their contents into the build structure.
// will panic if failure.
func (c *BuildConfig) ImportLanguages() {
	c.runJob(
		"languages",
		c.Folder.Languages,
		"/assets/lang/",
		true,
		[]string{".json"},
		func(l *log.Logger, f utilfunc.File, newPath string) {

			// read the JSON file
			langRaw, err := utilfunc.ReadJSONFile(f.Path, &Language{})
			if err != nil {
				l.Panicf("failed to read JSON file from OS [err: %v]", err)
			}

			// parse the JSON file
			lang := langRaw.(*Language)
			l.Printf("sucessfully parsed language [name: %v]", lang.LongName)

			c.languages = append(c.languages, *lang)

			// marshal JSON
			newJSON, err := sonic.Marshal(lang)
			if err != nil {
				l.Panicf("failed to marshal language JSON [err: %v]", err)
			}

			// write lang into distribution
			err = utilfunc.WriteFile(l, newPath, &newJSON)
			if err != nil {
				panic(err)
			}

		},
	)
}

// CopyStatic
// list all files on the passed static folder and
// copy all of them to the root of the distribution folder.
// will panic if failure.
func (c *BuildConfig) CopyStatic() {
	c.runJob(
		"static",
		c.Folder.Static,
		"/",
		true,
		[]string{},
		func(l *log.Logger, f utilfunc.File, newPath string) {
			err := utilfunc.CopyFile(l, f.Path, newPath)
			if err != nil {
				panic(err)
			}
		},
	)
}

// CopyUnhandableAssets
// copy to the distribution assets folder those files
// that can not be minified or processed, like images, svg
// files and fonts, for example. Keep the same file structure.
// will panic if failure.
func (c *BuildConfig) CopyUnhandableAssets() {
	c.runJob(
		"assets",
		c.Folder.Assets,
		"/assets/",
		false,
		[]string{".css", ".js", ".json"},
		func(l *log.Logger, f utilfunc.File, newPath string) {
			err := utilfunc.CopyFile(l, f.Path, newPath)
			if err != nil {
				panic(err)
			}
		},
	)
}

// MinifyAssets
// select the CSS, JS and JSON files at the assets source
// folder. Then, minify their contents and place the new,
// smaller version, into the distribution assets folder.
// will panic if failure.
func (c *BuildConfig) MinifyAssets() {
	c.runJob(
		"assets",
		c.Folder.Assets,
		"/assets/",
		true,
		[]string{".css", ".js", ".json"},
		func(l *log.Logger, f utilfunc.File, newPath string) {

			// determine the MIME kind
			mimekind := ""

			if strings.HasSuffix(newPath, ".js") {
				mimekind = "text/javascript"
			} else if strings.HasSuffix(newPath, ".json") {
				mimekind = "application/json"
			} else {
				mimekind = "text/css"
			}

			l.Printf("determined the file MIME kind [MIME: %v]", mimekind)

			// generate a minifier
			m := minify.New()
			switch mimekind {
			case "text/css":
				m.AddFunc("text/css", css.Minify)
			case "text/javascript":
				m.AddFunc("text/javascript", js.Minify)
			case "application/json":
				m.AddFunc("application/json", json.Minify)
			}

			// read the file
			sourceFile, err := os.Open(f.Path)
			if err != nil {
				l.Panicf("failed to read asset file from OS [err: %v]", err)
			}

			defer sourceFile.Close()

			l.Println("source file opened successfully")

			// do the minify operation
			destFile := bytes.NewBuffer([]byte{})

			if err := m.Minify(mimekind, destFile, sourceFile); err != nil {
				l.Panicf("failed to minify file [err: %v]", err)
			}

			l.Println("sucessfully minified file!")

			// write the file
			contents := destFile.Bytes()

			err = utilfunc.WriteFile(l, newPath, &contents)
			if err != nil {
				panic(err)
			}

		},
	)
}

// RenderViews
// will take the HTML files at the passed views folder
// and render them via the base view file. custom attributes
// for rendering might be passed. the result file, will be added to the
// root of the distribution folder, separated by each language
// short code.
// will panic if failure.
func (c *BuildConfig) RenderViews() {

	// setup a logger for this job
	lg := log.New(c.IODebug, "[job: html] ", log.LstdFlags|log.Lmsgprefix)
	lg.Println("starting compile procedures...")

	// list the files at the views folder
	files, err := utilfunc.ListFolderFiles(c.Folder.Views)
	if err != nil {
		lg.Panicf("failed to list folder files [err: %v]", err)
	}

	// generate a list of just HTML file paths
	var templatableFiles []string
	for _, f := range files {
		if f.IsDir || !strings.HasSuffix(f.Path, ".html") {
			continue
		}

		templatableFiles = append(templatableFiles, f.Path)
	}

	// !! minify

	// render the HTML files
	for _, path := range templatableFiles {

		// skip the base template
		if strings.HasSuffix(path, c.BaseView) {
			continue
		}

		// assemble a new logger for this file
		l := log.New(lg.Writer(), fmt.Sprintf("%v(%v) ", lg.Prefix(), path), lg.Flags())

		// rebase the file path
		newPath := strings.Replace(filepath.ToSlash(path), c.Folder.Views, c.DistributionPath+"/", 1)

		l.Printf("rebased file path [rebased: %v]", newPath)

		// assemble a new template
		tpl, err := template.New(path).ParseFiles(c.Folder.Views+c.BaseView, path)
		if err != nil {
			l.Panicf("failed to parse templates [err: %v]", err)
		}

		// !! execute the template
		renderedFile := bytes.NewBuffer([]byte{})
		err = tpl.ExecuteTemplate(renderedFile, "base", struct {
			Version string
			T       map[string]string
		}{
			Version: c.Version,
			T:       c.languages[0].Translations,
		})

		if err != nil {
			l.Panicf("failed to execute template [err: %v]", err)
		}

		l.Println("file rendered sucessfully")

		// write the file
		content := renderedFile.Bytes()

		err = utilfunc.WriteFile(l, newPath, &content)
		if err != nil {
			panic(err)
		}

	}

}

// UpdateWorker
// will list all the files at the distribution folder, then,
// get the worker.js file from the static folder, replace the variables
// at the worker file for use by the JS as a ServiceWorker.
// at the end, overwrite the worker.js file at the distribution folder.
// will panic if failure.
func (c *BuildConfig) UpdateWorker() {

	// setup a logger for this job
	lg := log.New(c.IODebug, "[job: worker] ", log.LstdFlags|log.Lmsgprefix)
	lg.Println("starting the worker.js compile procedures...")

	// read all the files from the dist directory
	files, err := utilfunc.ListFolderFiles(c.DistributionPath)
	if err != nil {
		lg.Panicf("failed to list folder files [err: %v]", err)
	}

	var paths []string
	for _, f := range files {
		if !f.IsDir {
			path := strings.Replace(filepath.ToSlash(f.Path), filepath.ToSlash(c.DistributionPath), "", 1)

			if path == "/version.txt" {
				continue
			}

			paths = append(paths, path)
		}
	}

	// join the file names into JS format
	toCache := strings.Join(paths, "',\n'")

	lg.Println("assembled the JS format")

	// read the JS file into a string
	dataRaw, err := os.ReadFile(c.Folder.Static + "/worker.js")
	if err != nil {
		lg.Panicf("failed to read worker file into string [err: %v]", err)
	}

	// update the worker file values
	data := string(dataRaw)
	data = strings.ReplaceAll(data, "{{tocache}}", toCache)
	data = strings.ReplaceAll(data, "{{version}}", c.Version)
	dataBytes := []byte(data)

	// write lang into distribution
	lg.Println("writing JS worker to file...")

	err = utilfunc.WriteFile(lg, c.DistributionPath+"/worker.js", &dataBytes)
	if err != nil {
		panic(err)
	}

}

// WriteVersionFile
// create a file "version.txt" at the distribution folder with the
// build version.
// will panic if failure.
func (c *BuildConfig) WriteVersionFile() {
	lg := log.New(c.IODebug, "[job: versionFile] ", log.LstdFlags|log.Lmsgprefix)
	dataBytes := []byte(c.Version)

	err := utilfunc.WriteFile(lg, c.DistributionPath+"/version.txt", &dataBytes)
	if err != nil {
		lg.Panicf("failed to write the version file [err: %v]", err)
	}
}
