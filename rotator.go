package rotator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	timeFormat = "2006-01-02T15-04-05.000000"
)

// Rotator represents a struct responsible for writing into a log file while
// rotating the file when it reached maxSize.
type Rotator struct {
	path        string
	filename    string
	currentSize uint64
	maxSize     uint64
	filesToKeep uint8
	writer      *os.File
	permissions os.FileMode
	mut         sync.Mutex
	useMut      bool
}

// New returns a new rotator prepared to be written to. The rotator is NOT thread-safe by default, since
// most logging libraries already take care of that.
//
// path: the path where log files should be written to, e.g. "/var/logs/myapp" or `C:\Logs`
//
// filename: the name the log files are supposed to have, e.g. 'test.log'
//
// maxSize: the maximum size in bytes a created log file may reach before it is rotated, e.g. 10 << 20 for 10MB
//
// perms: the file permissions in octal notation, e.g. 0744 (not relevant for windows)
//
// filesToKeep: the number of rotated files to keep. The currently written file is not counted towards this limit
func New(path, filename string, maxSize uint64, perms fs.FileMode, filesToKeep uint8, useMutex bool) (*Rotator, error) {
	r := Rotator{
		path:        path,
		filename:    filename,
		maxSize:     maxSize,
		permissions: perms,
		filesToKeep: filesToKeep,
		useMut:      useMutex,
	}

	err := os.MkdirAll(r.path, r.permissions)
	if err != nil {
		return nil, err
	}

	if stat, err := os.Stat(filepath.Join(r.path, r.filename)); err == nil {
		r.currentSize = uint64(stat.Size())
		if size := stat.Size(); size > int64(r.maxSize) {
			if r.writer != nil {
				r.writer.Close()
			}
			err = os.Rename(filepath.Join(r.path, r.filename), r.determineNextFilename())
			if err != nil {
				return nil, err
			}
			r.currentSize = 0
		}
	} else {
		if errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}

	fh, err := os.OpenFile(filepath.Join(r.path, r.filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, r.permissions)
	if err != nil {
		return nil, err
	}
	r.writer = fh

	return &r, nil
}

// Write writes the data into the log file and initiates rotation, if necessary
func (r *Rotator) Write(data []byte) (int, error) {
	if r.useMut {
		r.mut.Lock()
		defer r.mut.Unlock()
	}

	if r.currentSize+uint64(len(data)) > r.maxSize {
		if r.writer != nil {
			r.writer.Close()
		}

		err := r.removeUnnecessaryFiles()
		if err != nil {
			return 0, nil
		}

		err = os.Rename(filepath.Join(r.path, r.filename), r.determineNextFilename())
		if err != nil {
			return 0, err
		}
		fh, err := os.OpenFile(filepath.Join(r.path, r.filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, r.permissions)
		if err != nil {
			return 0, err
		}
		r.writer = fh
		r.currentSize = 0
	}
	r.currentSize += uint64(len(data))

	return r.writer.Write(data)
}

// determineNextFilename constructs the next filename to be used on rotation
func (r *Rotator) determineNextFilename() string {
	return filepath.Join(r.path, fmt.Sprintf("%s.%s", r.filename, time.Now().Format(timeFormat)))
}

// removeUnnecessaryFiles removes old files and keeps r.filesToKeep files
func (r *Rotator) removeUnnecessaryFiles() error {
	if r.filesToKeep == 0 {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(r.path, r.filename) + ".*")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		partsI := strings.Split(files[i], ".")
		partsJ := strings.Split(files[j], ".")
		dtI, err := time.Parse(timeFormat, partsI[len(partsI)-1])
		if err != nil {
			return false
		}
		dtJ, err := time.Parse(timeFormat, partsJ[len(partsJ)-1])
		if err != nil {
			return false
		}
		return dtI.Before(dtJ)
	})

	len := len(files)
	keep := int(r.filesToKeep)

	if keep >= len {
		return nil
	}

	filesToRemove := files[:len-keep]

	for _, f := range filesToRemove {
		err = os.Remove(filepath.Join(r.path, f))
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the io.Writer of the Rotator.
func (r *Rotator) Close() error {
	return r.writer.Close()
}
