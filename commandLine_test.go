package main

import (
	"os"
	"strings"
	"testing"

	"github.com/expel-io/cloud-resource-counter/mock"
)

func TestCommandLineProcess(t *testing.T) {
	// Our temp file
	const tempFile = "temp-output-file"

	// Construct our test cases...
	cases := []struct {
		Args             []string
		ExpectError      bool
		ExpectExit       bool
		ExpectAppend     bool
		ExpectAllRegions bool
		ExpectSSO        bool
	}{
		{
			Args:             []string{"--output-file", tempFile},
			ExpectAllRegions: true,
		},
		{
			Args:         []string{"--region", "us-west-2", "--output-file", tempFile},
			ExpectAppend: true,
		},
		{
			Args:             []string{"--output-file", tempFile, "--no-output"},
			ExpectError:      true,
			ExpectAllRegions: true,
		},
		{
			Args:        []string{"--region", "abc-def"},
			ExpectError: true,
		},
		{
			Args:             []string{"--version"},
			ExpectExit:       true,
			ExpectAllRegions: true,
		},
		{
			Args:             []string{"--sso", "--output-file", tempFile},
			ExpectAppend:     true,
			ExpectAllRegions: true,
			ExpectSSO:        true,
		},
	}

	// Does the file exist?
	if FileExists(tempFile) {
		// Remove the file
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Unexpected error while trying to delete temporary file: %v", err)
		}
	}

	// Loop through the cases...
	for _, c := range cases {
		// Create a Command Line
		settings := &CommandLineSettings{}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke the Process method
		cleanupFn := settings.Process(c.Args, mon)

		// Invoke the cleanup fn
		cleanupFn()

		// Did we expect an error?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else if mon.ProgramExited && !c.ExpectExit {
			t.Errorf("Unexpected Exit: The program unexpected exited with status code=%d", mon.ExitCode)
		} else if c.ExpectAppend != settings.appendToOutput {
			t.Errorf("Unexpected Append: expected %v, actual: %v", c.ExpectAppend, settings.appendToOutput)
		} else if c.ExpectAllRegions != settings.allRegions {
			t.Errorf("Unexpected AllRegions: expected %v, actual: %v", c.ExpectAllRegions, settings.allRegions)
		} else if c.ExpectSSO != settings.useSSO {
			t.Errorf("Unexpected SSO: expected %v, actual: %v", c.ExpectSSO, settings.useSSO)
		}
	}

	// Does the file exist?
	if FileExists(tempFile) {
		// Remove the file
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Unexpected error while trying to delete temporary file: %v", err)
		}
	}
}

func TestCommandLineTrace(t *testing.T) {
	// Our temp file
	const tempFile = "temp-trace-file"

	// Create a Command Line
	settings := &CommandLineSettings{}

	// Create a mock activity monitor
	mon := &mock.ActivityMonitorImpl{}

	// Invoke the Process method
	cleanupFn := settings.Process([]string{"--trace-file", tempFile, "--no-output"}, mon)

	// Invoke the cleanup fn
	cleanupFn()

	if mon.ErrorOccured {
		t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
	} else if mon.ProgramExited {
		t.Errorf("Unexpected Exit: The program unexpected exited with status code=%d", mon.ExitCode)
	} else if NilInterface(settings.traceFile) {
		t.Errorf("Expected traceFile to be written to, but it was not.")
	} else {
		// Remove the file
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Unexpected error while trying to delete temporary file: %v", err)
		}
	}
}

func TestCommandLineDisplay(t *testing.T) {
	// Create some test cases...
	cases := []struct {
		RegionName      string
		OutputFileName  string
		ExpectedStrings []string
		TraceFileName   string
	}{
		{
			ExpectedStrings: []string{"(All regions supported by this account)", "(none)"},
		},
		{
			RegionName:      "us-east-1",
			OutputFileName:  "bingo-pajamas",
			TraceFileName:   "trace-file",
			ExpectedStrings: []string{"us-east-1", "bingo-pajamas", "trace-file"},
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a Command Line
		settings := &CommandLineSettings{
			regionName:     c.RegionName,
			outputFileName: c.OutputFileName,
			traceFileName:  c.TraceFileName,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke the Display method}
		settings.Display(mon)

		// Ensure that we have some number of messages generated (don't want to get too tightly bound to impl)
		if len(mon.Messages) == 0 {
			t.Error("Expected to have some messages generated, but found none!")
		} else {
			// Loop through the list of expected strings
			var matchedStrings int
			for _, expectedString := range c.ExpectedStrings {
				// Loop through all of the messages
				for _, msg := range mon.Messages {
					// Is the expected string contained in the message?
					if strings.Contains(msg, expectedString) {
						matchedStrings++
					}
				}
			}

			// Do we have the expected number of strings?
			if matchedStrings != len(c.ExpectedStrings) {
				t.Errorf("Did not find all of our expected strings in the output: expected %d, actual %d", len(c.ExpectedStrings), matchedStrings)
			}
		}

		// Ensure that we didn't exit or error
		if mon.ErrorOccured || mon.ProgramExited {
			t.Errorf("Unexpected error (%v) or program exit (%v)", mon.ErrorOccured, mon.ProgramExited)
		}
	}
}
