package main

import (
	stdlog "log"
	"os"
)

//-----------------------------------------------------------------------------

// constants
const (
	// default file names
	paramFN = "param.go"
	defgnFN = "defgn.go"
	speclFN = "specl.go"
)

// default implementation
const (
	paramFT = `package %v

type T%v = interface{}
type U%v = interface{}
` // needs package-name,type-name,type-name

	defgnFT = `package %v

type %v map[T%v]U%v
` // needs package-name,type-name,type-name,type-name

	speclFT = `package %v
` // needs package-name
)

//-----------------------------------------------------------------------------

// errors
var (
	errNotDir  = errorf("NOT A DIR")
	errNotFile = errorf("NOT A FILE")
)

//-----------------------------------------------------------------------------

// global read-only
var (
	logerr = stdlog.New(os.Stderr, "err: ", 0)
	loginf = stdlog.New(os.Stdout, "inf: ", 0)
	logwrn = stdlog.New(os.Stdout, "wrn: ", 0)

	conf struct {
		Create struct {
			Name string `usage:"name of the generic type (required)" envvar:"-" name:"name,n"`
		}
	}
)

func init() {
	const dgb = false
	if dgb {
		logerr = stdlog.New(os.Stderr, "err: ", stdlog.Ltime|stdlog.Lshortfile)
		loginf = stdlog.New(os.Stdout, "inf: ", stdlog.Ltime|stdlog.Lshortfile)
		logwrn = stdlog.New(os.Stdout, "wrn: ", stdlog.Ltime|stdlog.Lshortfile)
	}
}

//-----------------------------------------------------------------------------
