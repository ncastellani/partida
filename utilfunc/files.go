package utilfunc

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

type File struct {
	Path    string
	IsDir   bool
	ModTime time.Time
	Name    string
	Size    int64
}

// ListFolderFiles
// take a folder path and walk through each file on the folder.
func ListFolderFiles(folder string) (files []File, err error) {

	// walk through each file on the folder
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// append this file/folder into the list
		files = append(files, File{
			Path:    path,
			IsDir:   info.IsDir(),
			ModTime: info.ModTime(),
			Name:    info.Name(),
			Size:    info.Size(),
		})

		return nil
	})

	return
}

// ReadJSONFile
// open an file within the passed path. then, try to
// unmarshal the obtained contents into the passed interface.
func ReadJSONFile(path string, parseTo interface{}) (parsed interface{}, err error) {

	// read the file from the OS
	file, err := os.Open(path)
	if err != nil {
		return
	}

	defer file.Close()

	// extract contents from the file
	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	// parse the JSON data into the interface
	err = sonic.Unmarshal(data, &parseTo)
	if err != nil {
		return
	}

	return parseTo, err
}

// CopyFile
// open and read the source file contents, then, call
// the WriteFile function to write the new file on the destination path.
func CopyFile(l *log.Logger, source, dest string) (err error) {
	l.Println("copying file...")

	// read the file contents
	file, err := os.Open(source)
	if err != nil {
		l.Printf("failed to read file from OS [err: %v]", err)
		return
	}

	defer file.Close()

	l.Println("sucessfully opened the source file")

	// write the file into the new path
	contents, err := io.ReadAll(file)
	if err != nil {
		l.Printf("failed to read file contents with io.ReadAll [err: %v]", err)
		return
	}

	return WriteFile(l, dest, &contents)
}

// WriteFile
// take an path to write into and content in
// bytes, and write the file creating the folder tree recursively.
func WriteFile(l *log.Logger, path string, content *[]byte) (err error) {
	l.Println("writing file...")

	// determine the folder structure
	pathSplitted := strings.Split(path, "/")
	folder := strings.Join(pathSplitted[:len(pathSplitted)-1], "/")

	l.Printf("determined the file tree [folderTree: %v]", folder)

	// create the folder if not exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		l.Println("folder does not exists. creating...")

		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// open the destination file
	destFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		l.Printf("failed to open the destination file [err: %v]", err)
		return
	}

	defer destFile.Close()

	l.Println("destination file opened successfully")

	// write into the file
	_, err = destFile.Write(*content)
	if err != nil {
		l.Printf("failed to write file on disk [err: %v]", err)
		return
	}

	l.Println("the file was written sucessfully")

	return
}
