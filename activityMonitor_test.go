package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

func TestTerminalActivityMonitorMessage(t *testing.T) {
	// Create a builder to hold our contents...
	builder := strings.Builder{}

	// Create an instance of the Terminal Activity Monitor
	mon := TerminalActivityMonitor{
		Writer: &builder,
	}

	// Send it some messages...
	mon.Message("Testing %d,%d,%d", 1, 2, 3)
	if builder.String() != "Testing 1,2,3" {
		t.Errorf("Unexpected message: expected %s, actual %s", "Testing 1,2,3", builder.String())
	}
}

func TestTerminalActivityMonitorStartAction(t *testing.T) {
	// Create a builder to hold our contents...
	builder := strings.Builder{}

	// Create an instance of the Terminal Activity Monitor
	mon := TerminalActivityMonitor{
		Writer: &builder,
	}

	// Start an action
	mon.StartAction("Doing something good")
	if builder.String() != " * Doing something good..." {
		t.Errorf("Unexpected StartAction: expected %s, actual %s", " * Doing something good...", builder.String())
	}
}

func TestTerminalActivityMonitorEndAction(t *testing.T) {
	// Create a builder to hold our contents...
	builder := strings.Builder{}

	// Create an instance of the Terminal Activity Monitor
	mon := TerminalActivityMonitor{
		Writer: &builder,
	}

	// Start an action
	mon.EndAction("OK (%d)", 123)
	if builder.String() != "OK (123)\n" {
		t.Errorf("Unexpected EndAction: expected %s, actual %s", "OK (123)\n", builder.String())
	}
}

func TestTerminalActivityMonitorCheckError(t *testing.T) {
	// Create some test cases with actual errors...
	cases := []struct {
		Error          error
		IsEmpty        bool
		ContainsString string
	}{
		{
			IsEmpty: true,
		},
		{
			Error:          errors.New("Something is very wrong"),
			ContainsString: "Something is very wrong",
		},
		{
			Error:          awserr.New("NoCredentialProviders", "blah", nil),
			ContainsString: "profile does not exist",
		},
		{
			Error:          awserr.New("AccessDeniedException", "You don't have access to do this stuff\nAnd another thing", nil),
			ContainsString: "You don't have access to do this stuff",
		},
		{
			Error:          awserr.New("InvalidClientTokenId", "This is a very technical error\nBlah, blah, blah", nil),
			ContainsString: "region is not supported",
		},
		{
			Error:          awserr.New("SomethingSmellsFishy", "There is something not quite right\nNot really sure what it is", nil),
			ContainsString: "SomethingSmellsFishy: There is something not quite right",
		},
	}

	// Loop through the test cases...
	for _, c := range cases {
		// Create an exit function which simply sets a value
		var exitMsg string
		exitFn := func(resultCode int) {
			exitMsg = fmt.Sprintf("Status %d", resultCode)
		}

		// Create a builder to hold our contents...
		builder := strings.Builder{}

		// Create an instance of the Terminal Activity Monitor
		mon := TerminalActivityMonitor{
			Writer: &builder,
			ExitFn: exitFn,
		}

		// Check for error
		mon.CheckError(c.Error)

		// Get our actual string
		actual := builder.String()

		// Do we expect a string?
		if c.IsEmpty && actual != "" {
			t.Errorf("Unexpected CheckError: expected %s, actual %s", "", actual)
		} else if !strings.Contains(actual, c.ContainsString) {
			t.Errorf("Unexpected CheckError: expected '%s' to be contained in '%s', but it was not", c.ContainsString, actual)
		} else if c.Error != nil && exitMsg != "Status 1" {
			t.Errorf("Unexpected CheckError: ExitFn: expected %s, actual %s", "Status 1", exitMsg)
		}
	}
}
