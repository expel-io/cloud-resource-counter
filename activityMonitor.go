package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
	color "github.com/logrusorgru/aurora"
)

// ActivityMonitor is an interface that shows activity to the user while
// the tool runs. It handles the start of an action, checking for an error,
// displaying the error or the actual result.
type ActivityMonitor interface {
	Message(string, ...interface{})
	StartAction(string, ...interface{})
	CheckError(error)
	ActionError(string, ...interface{})
	EndAction(string, ...interface{})
}

// TerminalActivityMonitor is our terminal-based activity monitor. It allows
// the caller to supply an io.Writer to direct output to.
type TerminalActivityMonitor struct {
	io.Writer
}

// Message constructs a simple message from the format string and arguments
// and sends it to the associated io.Writer.
func (tam *TerminalActivityMonitor) Message(format string, v ...interface{}) {
	fmt.Fprint(tam.Writer, fmt.Sprintf(format, v...))
}

// StartAction constructs a structured message to the associated Writer
func (tam *TerminalActivityMonitor) StartAction(format string, v ...interface{}) {
	tam.Message(" * %s...", fmt.Sprintf(format, v...))
}

// CheckError checks the supplied error. If no error, then it returns immediately.
// Otherwise, it checks for specific AWS errors (returning a specific error message).
// If no specific AWS error found, it simply sends the error message to the ActionError
// method.
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
			// Construct our message based on whether a profile name was specified
			var message string
			if profileName != defaultProfileName {
				message = "Either the profile name is misspelled or credentials are not stored there."
			} else {
				message = "Either the default profile does not exist or credentials are not stored there."
			}
			tam.ActionError(message)
			break
		default:
			tam.ActionError("%v", aerr)
		}
	} else {
		tam.ActionError("%v", err)
	}
}

// ActionError formats the supplied format string (and associated parameters) in
// RED and exits the tool.
func (tam *TerminalActivityMonitor) ActionError(format string, v ...interface{}) {
	fmt.Fprintln(tam.Writer, color.Red(fmt.Sprintf(format, v...)))
	fmt.Fprintln(tam.Writer)

	os.Exit(1)
}

// EndAction receives a format string (and arguments) and sends to the supplied
// Writer.
func (tam *TerminalActivityMonitor) EndAction(format string, v ...interface{}) {
	fmt.Fprintln(tam.Writer, fmt.Sprintf(format, v...))
}
