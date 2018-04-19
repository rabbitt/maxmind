// Largely borrowed from https://github.com/mitchellh/cli/blob/master/ui.go

package command

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Ui is an interface for interacting with the terminal, or "interface"
// of a CLI. This abstraction doesn't have to be used, but helps provide
// a simple, layerable way to manage user interactions.
type Ui interface {
	// Output is called for normal standard output.
	Output(...interface{})
	Outputf(string, ...interface{})

	// Info is called for information related to the previous output.
	// In general this may be the exact same as Output, but this gives
	// Ui implementors some flexibility with output formats.
	Info(...interface{})
	Infof(string, ...interface{})

	// Error is used for any error messages that might appear on standard
	// error.
	Error(...interface{})
	Errorf(string, ...interface{})

	// Fatal are like Error, but exit immediately after rendering their message
	Fatal(...interface{})
	Fatalf(string, ...interface{})

	// Warn is used for any warning messages that might appear on standard
	// error.
	Warn(...interface{})
	Warnf(string, ...interface{})
}

// BaseUi is an implementation of Ui that just outputs to the given
// writer. This UI is threadsafe by default.

type BaseUi struct {
	Writer      io.Writer
	ErrorWriter io.Writer
	locker      sync.Mutex
}

func (u *BaseUi) Error(v ...interface{}) {
	w := u.Writer
	if u.ErrorWriter != nil {
		w = u.ErrorWriter
	}

	u.locker.Lock()
	defer u.locker.Unlock()

	fmt.Fprint(w, v...)
	fmt.Fprint(w, "\n")
}

func (u *BaseUi) Errorf(format string, v ...interface{}) {
	w := u.Writer
	if u.ErrorWriter != nil {
		w = u.ErrorWriter
	}

	u.locker.Lock()
	defer u.locker.Unlock()

	fmt.Fprintf(w, format, v...)
}

func (u *BaseUi) Fatal(v ...interface{}) {
	u.Error(v...)
	os.Exit(1)
}

func (u *BaseUi) Fatalf(format string, v ...interface{}) {
	u.Errorf(format, v...)
	os.Exit(1)
}

func (u *BaseUi) Info(v ...interface{}) {
	u.Output(v...)
}

func (u *BaseUi) Infof(format string, v ...interface{}) {
	u.Outputf(format, v...)
}

func (u *BaseUi) Output(v ...interface{}) {
	u.locker.Lock()
	defer u.locker.Unlock()

	fmt.Fprint(u.Writer, v...)
	fmt.Fprint(u.Writer, "\n")
}

func (u *BaseUi) Outputf(format string, v ...interface{}) {
	u.locker.Lock()
	defer u.locker.Unlock()

	fmt.Fprintf(u.Writer, format, v...)
}

func (u *BaseUi) Warn(v ...interface{}) {
	u.Error(v...)
}

func (u *BaseUi) Warnf(format string, v ...interface{}) {
	u.Errorf(format, v...)
}

// PrefixedUi is an implementation of Ui that prefixes messages.
type PrefixedUi struct {
	Ui Ui
}

func (u *PrefixedUi) PrependPrefix(prefix string, args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = prefix
	for idx, value := range args {
		newArgs[idx+1] = value
	}
	return newArgs
}

func (u *PrefixedUi) Error(v ...interface{}) {
	u.Ui.Error(u.PrependPrefix("ERROR: ", v)...)
}

func (u *PrefixedUi) Errorf(format string, v ...interface{}) {
	u.Ui.Errorf(fmt.Sprintf("ERROR: %s", format), v...)
}

func (u *PrefixedUi) Fatal(v ...interface{}) {
	u.Ui.Fatal(u.PrependPrefix("FATAL: ", v)...)
}

func (u *PrefixedUi) Fatalf(format string, v ...interface{}) {
	u.Ui.Fatalf(fmt.Sprintf("FATAL: %s", format), v...)
}

func (u *PrefixedUi) Info(v ...interface{}) {
	u.Ui.Info(u.PrependPrefix("INFO: ", v)...)
}

func (u *PrefixedUi) Infof(format string, v ...interface{}) {
	u.Ui.Infof(fmt.Sprintf("INFO: %s", format), v...)
}

func (u *PrefixedUi) Output(v ...interface{}) {
	u.Ui.Output(v...)
}

func (u *PrefixedUi) Outputf(format string, v ...interface{}) {
	u.Ui.Outputf(format, v...)
}

func (u *PrefixedUi) Warn(v ...interface{}) {
	u.Ui.Warn(u.PrependPrefix("WARN: ", v)...)
}

func (u *PrefixedUi) Warnf(format string, v ...interface{}) {
	u.Ui.Warnf(fmt.Sprintf("WARN: %s", format), v...)
}
