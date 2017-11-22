package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/tonistiigi/fifo"
)

// LoggingDriver defines the interface for types that want to be a Docker logging
// plugin for the DE.
type LoggingDriver interface {
	StartLogging(file string, info logger.Info) error
	StopLogging(file string) error
}

// FakeDriver doesn't actually do anything except log when it receives a message.
type FakeDriver struct{}

// StartLogging doesn't do anything except log that it was called.
func (f *FakeDriver) StartLogging(file string, info logger.Info) error {
	log.Printf("StartLogging called with file %s for container %s\n", file, info.ContainerID)
	return nil
}

// StopLogging doesn't do anything except log that it was called.
func (f *FakeDriver) StopLogging(file string) error {
	log.Printf("StopLogging was called with file %s\n", file)
	return nil
}

// FileLogger tracks the info needed to write out log messages to files.
type FileLogger struct {
	StderrPath string
	StdoutPath string
	Stdout     *os.File
	Stderr     *os.File
	LogStream  io.ReadCloser
}

// StreamMessages will consume logging messages sent from Docker to the FIFO
// stream and write them out to the configured log files.
func (l *FileLogger) StreamMessages() {
	reader := protoio.NewUint32DelimitedReader(l.LogStream, binary.BigEndian, 1e6)
	defer reader.Close()

	var (
		err   error
		entry logdriver.LogEntry
	)

	for {
		if err = reader.ReadMsg(&entry); err != nil {
			if err == io.EOF {
				l.LogStream.Close()
				return
			}
			reader = protoio.NewUint32DelimitedReader(l.LogStream, binary.BigEndian, 1e6)
		}

		msg := logger.Message{
			Line:      append(entry.Line, []byte("\n")...),
			Source:    entry.Source,
			Partial:   entry.Partial,
			Timestamp: time.Unix(0, entry.TimeNano),
		}

		switch msg.Source {
		case "stderr":
			if _, err = l.Stderr.Write(msg.Line); err != nil {
				err = errors.Wrap(err, "error writing to stderr log file")
				log.Println(err.Error())
				continue
			}
		case "stdout":
			if _, err = l.Stdout.Write(msg.Line); err != nil {
				err = errors.Wrap(err, "error writing to stdout log file")
				log.Println(err.Error())
				continue
			}
		default:
			log.Println(fmt.Errorf("Unknown source %s for message: %s", msg.Source, msg.Line))
			continue
		}

		entry.Reset()
	}
}

// FileDriver is a logging driver that will write out the stderr and stdout
// streams to a configured directory.
type FileDriver struct {
	mu     sync.Mutex
	logmap map[string]*FileLogger
	base   string
}

// NewFileDriver returns a newly created *FileDriver.
func NewFileDriver() (*FileDriver, error) {
	basepath := "/var/log/de-docker-logging-plugin"
	_, err := os.Stat(basepath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(basepath, 0755); err != nil {
			return nil, errors.Wrapf(err, "error creating directory %s", basepath)
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error stat'ing %s", basepath)
	}
	return &FileDriver{
		logmap: make(map[string]*FileLogger),
		base:   basepath,
	}, nil
}

// StartLogging sets up everything needed for logging to separate files for
// stdout and stderr. Fires up a goroutine that pipes info from the FIFO created
// by Docker into each file.
func (d *FileDriver) StartLogging(fifopath string, loginfo logger.Info) error {
	d.mu.Lock()
	if _, ok := d.logmap[fifopath]; ok {
		d.mu.Unlock()
		return fmt.Errorf("logging is already configured for %s", fifopath)
	}
	d.mu.Unlock()

	if _, ok := loginfo.Config["stderr"]; !ok {
		return fmt.Errorf("'stderr' path missing from the plugin configuration")
	}

	if _, ok := loginfo.Config["stdout"]; !ok {
		return fmt.Errorf("'stdout' path missing from the plugin configuration")
	}

	baseDir := d.base

	gid, err := strconv.Atoi(os.Getenv("gid"))
	if err != nil {
		return errors.Wrap(err, "failed to convert the gid env var to an int")
	}
	uid, err := strconv.Atoi(os.Getenv("uid"))
	if err != nil {
		return errors.Wrap(err, "failed to convert the uid env var to an int")
	}

	stderrPath := path.Join(baseDir, loginfo.Config["stderr"])
	stdoutPath := path.Join(baseDir, loginfo.Config["stdout"])

	stderrBase := path.Dir(stderrPath)
	stdoutBase := path.Dir(stdoutPath)

	for _, p := range []string{stderrBase, stdoutBase} {
		pinfo, err := os.Stat(p)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(p, 0755); err != nil {
				return errors.Wrapf(err, "error creating %s", p)
			}
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "error stat'ing path %s", p)
		}
		if !pinfo.IsDir() {
			return errors.Wrapf(err, "path was not a directory %s", p)
		}
	}

	for _, p := range []string{loginfo.Config["stderr"], loginfo.Config["stdout"]} {
		acc := baseDir
		parentdir := path.Dir(p)
		for _, d := range strings.Split(parentdir, string(os.PathSeparator)) {
			if d != "" {
				acc = path.Join(acc, d)
				if err = os.Chown(acc, uid, gid); err != nil {
					return errors.Wrapf(err, "failed to chown %s to %d:%d", acc, uid, gid)
				}
			}
		}
	}

	stderr, err := os.Create(stderrPath)
	if err != nil {
		return errors.Wrapf(err, "error opening stderr log file at %s", stderrBase)
	}
	if err = stderr.Chown(uid, gid); err != nil {
		return errors.Wrapf(err, "failed to chown %s to %d:%d", stderrPath, uid, gid)
	}

	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return errors.Wrapf(err, "error opening stdout log file at %s", stdoutBase)
	}
	if err = stdout.Chown(uid, gid); err != nil {
		return errors.Wrapf(err, "failed to chown %s to %d:%d", stdoutPath, uid, gid)
	}

	f, err := fifo.OpenFifo(context.Background(), fifopath, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening fifo file %s", fifopath)
	}

	filelogger := &FileLogger{
		StderrPath: stderrPath,
		StdoutPath: stdoutPath,
		Stderr:     stderr,
		Stdout:     stdout,
		LogStream:  f,
	}

	d.mu.Lock()
	d.logmap[fifopath] = filelogger
	d.mu.Unlock()

	go filelogger.StreamMessages()

	return nil
}

// StopLogging terminates logging to files and closes them out.
func (d *FileDriver) StopLogging(fifopath string) error {
	d.mu.Lock()
	fl, ok := d.logmap[fifopath]
	if ok {
		fl.LogStream.Close()
		fl.Stderr.Close()
		fl.Stdout.Close()
		delete(d.logmap, fifopath)
	}
	d.mu.Unlock()
	return nil
}
