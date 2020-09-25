package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
	color "github.com/logrusorgru/aurora"
)

// ActivityMonitor is the shit!
type ActivityMonitor interface {
	StartAction(string, ...interface{})
	CheckError(error)
	ActionError(string, ...interface{})
	EndAction(string, ...interface{})
}

// TerminalActivityMonitor is our default terminal-based activity monitor
type TerminalActivityMonitor struct{}

// StartAction is the beginning of an activity
func (tam *TerminalActivityMonitor) StartAction(format string, v ...interface{}) {
	// Send the start of an activity to STDERR
	fmt.Fprintf(os.Stderr, " * %s...", fmt.Sprintf(format, v...))
}

// CheckError does stuff...
func (tam *TerminalActivityMonitor) CheckError(err error) {
	// If it is nil, get out now!
	if err == nil {
		return
	}

	// Is this an AWS Error?
	if aerr, ok := err.(awserr.Error); ok {
		// Switch on the error code for known error conditions...
		switch aerr.Code() {
		case "NoCredentialProviders":
			// TODO Can we establish this failure earlier? When the session is created?
			tam.ActionError("Either the profile name is misspelled or credentials are not stored.")
			break
		default:
			tam.ActionError("%v", aerr)
		}
	} else {
		tam.ActionError("%v", err)
	}
}

// ActionError such a good name...
func (tam *TerminalActivityMonitor) ActionError(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, color.Red(fmt.Sprintf(format, v...)))
	fmt.Fprintln(os.Stderr)

	os.Exit(1)
}

// EndAction ...
func (tam *TerminalActivityMonitor) EndAction(format string, v ...interface{}) {
	// Send the start of an activity to STDERR
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
}
