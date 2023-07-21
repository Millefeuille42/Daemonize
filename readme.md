# Daemonize

A daemonization toolkit for golang processes

## Usage

Import the package

```go
package main

import "github.com/Millefeuille42/Daemonize"
```

Create a new Daemonizer

```go
package main

import (
	"github.com/Millefeuille42/Daemonize"
	"log"
)

func main() {
	d, err := Daemonize.NewDaemonizer()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()	
}
```

Start the process as a daemon

```go
package main

import (
	"github.com/Millefeuille42/Daemonize"
	"log"
)
func main() {
	d, err := Daemonize.NewDaemonizer()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	pid, err := d.Daemonize(nil)
	if err != nil {
		log.Fatal(err)
	}
}
```

This will start the program as a daemon, with the working directory set as system root
and with a logger interface on syslog.

Like fork, `Daemonize` returns a non 0 PID if the process is the parent process 
(child's PID for success, `1` for error) if the process is the child process, it returns `0`.

It is recommended to exit the parent process right after.

```go
package main

import (
	"github.com/Millefeuille42/Daemonize"
	"log"
	"os"
)

func main() {
	d, err := Daemonize.NewDaemonizer()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	pid, err := d.Daemonize(nil)
	if err != nil {
		log.Fatal(err)
	}
	if pid != 0 {
		// In parent process
		log.Print(pid)
		os.Exit(0)
	}
	// In child process
}
```

It is possible to add loggers thanks to various functions, example: 

```go
d.AddFileLogger("/path/to/log/file", os.Args[0], log.LstdFlags)
```

Logging is done via the `d.Log` function, it logs on the registered loggers and to syslog

```go
d.Log(syslog.LOG_INFO, "Hello there")
```

The severity parameters is for syslog.

The `d.Close` function also catches panic events to log it on files and syslog.

The "minimal" client code is as follows:

```go
package main

import (
	"github.com/Millefeuille42/Daemonize"
	"log"
	"os"
)

func main() {
	d, err := Daemonize.NewDaemonizer()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	pid, err := d.Daemonize(nil)
	if err != nil {
		log.Fatal(err)
	}
	if pid != 0 {
		log.Print(pid)
		os.Exit(0)
	}

	err = d.AddTempFileLogger("/path/to/log/folder", "daemon_*.log", os.Args[0], log.LstdFlags)
	if err != nil {
		log.Fatal(err)
	}

	d.Log(Daemonize.LOG_INFO, "Hello there")
}
```

## Definitions

### Severity

Severity is an indicator of the purpose of the log message, it is linked to syslog Priority levels.

```go
type Severity int

const (
    // LOG_DEBUG Useful data for debugging
    LOG_DEBUG Severity = iota
    // LOG_INFO Non-important information, considered to be the default level
    LOG_INFO
    // LOG_WARNING Rare or unexpected conditions
    LOG_WARNING
    // LOG_ERR Errors
    LOG_ERR
    // LOG_EMERG Fatal errors
    LOG_EMERG
)
```
