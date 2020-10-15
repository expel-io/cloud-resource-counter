/******************************************************************************
Cloud Resource Counter
File: commandLine.go

Summary: Retrieve account ID (assumed to be a single value) for the current
         user session.
******************************************************************************/

package main

import (
	"flag"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	color "github.com/logrusorgru/aurora"
)

// CommandLineSettings defines the command line settings supplied by
// the caller.
type CommandLineSettings struct {
	// Profile related settings
	profileName        string
	defaultProfileName string

	// Region related settings
	regionName string

	// Output (CSV) file
	outputFileName string
	outputFile     *os.File
	appendToOutput bool

	// Trace file
	traceFileName string
	traceFile     *os.File
}

// Process inspects the command line for valid arguments.
//
// Usage of cloud-resource-counter
//   --append:         Whether to append to an existing output file or not
//   --output-file OF: Write the results to file OF
//   --profile PN:     Use the credentials associated with shared profile PN
//   --region RN:      View resource counts for the AWS region RN
//   --trace-file TF:  Create a trace file that contains all calls to AWS.
//   --version:        Display version information
//
func (cls *CommandLineSettings) Process(am ActivityMonitor) {
	var showVersion bool

	// What is our default profile?
	if cls.defaultProfileName = os.Getenv("AWS_PROFILE"); cls.defaultProfileName == "" {
		cls.defaultProfileName = session.DefaultSharedConfigProfile
	}

	// Define and parse the command line arguments...
	flag.BoolVar(&cls.appendToOutput, "append", false, "Whether to append to an existing output file or not. (default false--replace previous contents)")
	flag.StringVar(&cls.outputFileName, "output-file", "./resources.csv", "CSV Output File. Specify a path to a `file` to save the generated CSV file.")
	flag.StringVar(&cls.profileName, "profile", cls.defaultProfileName, "The name of the AWS Profile to use.")
	flag.StringVar(&cls.regionName, "region", "", "The name of the AWS Region to use. If omitted, then all regions will be examined. This is the default behavior")
	flag.StringVar(&cls.traceFileName, "trace-file", "", "AWS Trace Log. Specify a `file` to record API calls being made.")
	flag.BoolVar(&showVersion, "version", false, "Shows the version number.")
	flag.Parse()

	// Check for a valid AWS Region
	if cls.regionName != "" {
		// If not valid region name, then get out now...
		if !IsValidRegionName(cls.regionName) {
			am.ActionError("Error: '%s' is not a valid AWS Region name.", cls.regionName)
		}
	}

	// Did the user just want to see the version?
	if showVersion {
		am.Message("%s, version %s (built %s)\n", "Cloud Resource Counter", version, date)
		os.Exit(0)
	}

	// Check whether a response file is being specified
	if cls.outputFileName != "" {
		// Try to open the file for writing
		cls.outputFile = OpenFileForWriting(cls.outputFileName, "CSV", am, cls.appendToOutput)
	}

	// Check whether a trace file is being specified
	if cls.traceFileName != "" {
		// Try to open the file for writing
		cls.traceFile = OpenFileForWriting(cls.traceFileName, "trace", am, false)
	}
}

// Display constructs a listing of all command line settings to the Activity Monitor
func (cls *CommandLineSettings) Display(am ActivityMonitor) {
	// What is the region being selected?
	var displayRegionName string
	if cls.regionName == "" {
		displayRegionName = "(All regions supported by this account)"
	} else {
		displayRegionName = cls.regionName
	}

	// Output information about utility running
	am.Message("%s (v%s) running with:\n", color.Bold("Cloud Resource Counter"), version)
	am.Message(" o %s: %s\n", color.Italic("AWS Profile"), cls.profileName)
	am.Message(" o %s:  %s\n", color.Italic("AWS Region"), displayRegionName)
	am.Message(" o %s: %s\n", color.Italic("Output file"), cls.outputFileName)

	// Are we tracing?
	if cls.traceFileName != "" {
		am.Message(" o %s:  %s\n", color.Italic("Trace file"), cls.traceFileName)
	}
}
