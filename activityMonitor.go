/******************************************************************************
Cloud Resource Counter
File: activityMonitor.go

Summary: The ActivityMonitor interface along with a Terminal-based monitor.
******************************************************************************/

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	color "github.com/logrusorgru/aurora"
)

// ActivityMonitor is an interface that shows activity to the user while
// the tool runs. It handles the start of an action, checking for an error,
// displaying the error or the actual result.
type ActivityMonitor interface {
	Message(string, ...interface{})
	StartAction(string, ...interface{})
	CheckError(error) bool
	ActionError(string, ...interface{})
	EndAction(string, ...interface{})
	Exit(int)
}

// TerminalActivityMonitor is our terminal-based activity monitor. It allows
// the caller to supply an io.Writer to direct output to.
type TerminalActivityMonitor struct {
	io.Writer
	ExitFn func(int)
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
func (tam *TerminalActivityMonitor) CheckError(err error) bool {
	// If it is nil, get out now!
	if err == nil {
		return false
	}

	// Is this an AWS Error?
	if aerr, ok := err.(awserr.Error); ok {
		// Split the message by newline
		parts := strings.Split(aerr.Message(), "\n")

		// Switch on the error code for known error conditions...
		switch aerr.Code() {
		case "NoCredentialProviders":
			// TODO Can we establish this failure earlier? When the session is created?
			tam.ActionError("Either the profile does not exist, is misspelled or credentials are not stored there.")
			break
		case "AccessDeniedException":
			// Construct a message by taking the first part of the string up to a newline character
			tam.ActionError(parts[0])
		case "InvalidClientTokenId":
			// Construct a message that indicates an unsupported region
			tam.ActionError("The region is not supported for this account.")
		default:
			tam.ActionError("%s: %s", aerr.Code(), parts[0])
		}
	} else {
		tam.ActionError("%v", err)
	}

	return true
}

// ActionError formats the supplied format string (and associated parameters) in
// RED and exits the tool.
func (tam *TerminalActivityMonitor) ActionError(format string, v ...interface{}) {
	// Display an error message (and newline)
	fmt.Fprintln(tam.Writer, color.Red(fmt.Sprintf(format, v...)))
	fmt.Fprintln(tam.Writer)

	// Exit the program
	tam.Exit(1)
}

// EndAction receives a format string (and arguments) and sends to the supplied
// Writer.
func (tam *TerminalActivityMonitor) EndAction(format string, v ...interface{}) {
	fmt.Fprintln(tam.Writer, fmt.Sprintf(format, v...))
}

// Exit causes the application to exit
func (tam *TerminalActivityMonitor) Exit(resultCode int) {
	if tam.ExitFn != nil {
		tam.ExitFn(resultCode)
	}
}
