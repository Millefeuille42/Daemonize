package Daemonize

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
)

// Daemonizer is a helper struct for "daemonizing" go processes
type Daemonizer struct {
	// SyslogWriter is the writer used to write on the syslog
	SyslogWriter *syslog.Writer
	// Loggers is a slice containing all registered loggers
	Loggers []*log.Logger
	// files is the list of opened file (with the logger system), used in Close to close the files
	files []*os.File
	// sid is the SID of the daemon
	sid int
}

// AddLogger adds a logger to the Loggers slice
func (d *Daemonizer) AddLogger(logger *log.Logger) {
	d.Loggers = append(d.Loggers, logger)
}

func (d *Daemonizer) addFileLogger(file *os.File, prefix string, flags int) {
	d.AddLogger(log.New(file, prefix, flags))
}

// AddFileLogger opens a writer to the file located at path and opens a logger on it, the adds it to the Loggers slice
func (d *Daemonizer) AddFileLogger(path, prefix string, flags int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	d.addFileLogger(file, prefix, flags)
	return nil
}

// AddTempFileLogger opens a writer to the file located at path and opens a logger on it,
// then adds it to the Loggers slice. It uses the os.CreateTemp function to open files
func (d *Daemonizer) AddTempFileLogger(dir, pattern, prefix string, flags int) error {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return err
	}
	d.addFileLogger(file, prefix, flags)
	return nil
}

// Log logs on SyslogWriter and Loggers, the severity is used for syslog
func (d *Daemonizer) Log(severity Severity, v ...any) {
	switch severity {
	case LOG_DEBUG:
		_ = d.SyslogWriter.Debug(fmt.Sprint(v...))
	case LOG_WARNING:
		_ = d.SyslogWriter.Warning(fmt.Sprint(v...))
	case LOG_ERR:
		_ = d.SyslogWriter.Err(fmt.Sprint(v...))
	case LOG_EMERG:
		_ = d.SyslogWriter.Emerg(fmt.Sprint(v...))
	default:
		_ = d.SyslogWriter.Info(fmt.Sprint(v...))
	}

	for _, logger := range d.Loggers {
		logger.Print(v...)
	}
}

// Sid returns the daemons SID
func (d *Daemonizer) Sid() int {
	return d.sid
}

// Daemonize uses os.StartProcess to start itself as a daemon, like fork, returns 0 if in the child process,
// returns child process Pid if spawned, returns 1 if an error occurred in the parent process, unlike fork.
// the args parameters corresponds to the arguments passed to the child process
// the caller name is already added
func (d *Daemonizer) Daemonize(args []string) (int, error) {
	sid, err := syscall.Setsid()
	if err != nil {
		self, err := os.Executable()
		if err != nil {
			return 1, err
		}

		pArgs := make([]string, 0)
		pArgs = append(pArgs, self)
		if args != nil {
			pArgs = append(pArgs, args...)
		}

		child, err := os.StartProcess(self, pArgs, &os.ProcAttr{})
		if err != nil {
			return 1, err
		}
		return child.Pid, err
	}

	d.sid = sid
	return 0, os.Chdir("/")
}

// Close closes SyslogWriter and all files. It also catches panic() events to log it in SyslogWriter and Loggers
func (d *Daemonizer) Close() error {
	var err error = nil
	if pnc := recover(); pnc != nil {
		d.Log(LOG_EMERG, pnc)
	}
	for _, file := range d.files {
		fErr := file.Close()
		if fErr != nil {
			err = fErr
		}
	}
	fErr := d.SyslogWriter.Close()
	if fErr != nil {
		err = fErr
	}
	d.Log(LOG_INFO, "Daemon closed")
	return err
}

func (d *Daemonizer) HandleSignals(additionalHandler func() error) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Kill, syscall.SIGTERM)
	if additionalHandler != nil {
		err := additionalHandler()
		if err != nil {
			d.Log(LOG_ERR, err)
		}
	}
	_ = d.Close()
}

// NewDaemonizer creates a new Daemonizer instance and creates a writer to syslog
func NewDaemonizer() (*Daemonizer, error) {
	d := &Daemonizer{
		Loggers: make([]*log.Logger, 0),
		files:   make([]*os.File, 0),
	}

	syslogWriter, err := syslog.New(syslog.LOG_USER, "")
	if err != nil {
		return nil, err
	}

	d.SyslogWriter = syslogWriter
	log.SetOutput(d.SyslogWriter)
	return d, nil
}
