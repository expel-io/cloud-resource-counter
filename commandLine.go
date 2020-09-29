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
	allRegions bool

	// Output (CSV) file
	outputFileName string
	outputFile     *os.File

	// Trace file
	traceFileName string
	traceFile     *os.File
}

// Process inspects the command line for valid arguments.
//
// Usage of cloud-resource-counter
//   --all-regions:    Collect counts for all regions associated with the account
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

	// What is our default region
	defaultRegionName := os.Getenv("AWS_REGION")

	// Define and parse the command line arguments...
	flag.BoolVar(&cls.allRegions, "all-regions", false, "Whether to iterate over all regions associated with the account.")
	flag.StringVar(&cls.outputFileName, "output-file", "./resources.csv", "CSV Output File. Specify a path to a file to save the generated CSV file")
	flag.StringVar(&cls.profileName, "profile", cls.defaultProfileName, "AWS Profile Name")
	flag.StringVar(&cls.regionName, "region", defaultRegionName, "Selects an AWS Region to use")
	flag.StringVar(&cls.traceFileName, "trace-file", "", "AWS Trace Log. Specify a file to record API calls being made.")
	flag.BoolVar(&showVersion, "version", false, "Shows the version number.")
	flag.Parse()

	// TODO Check for a valid AWS Region

	// Did the user just want to see the version?
	if showVersion {
		am.Message("%s, version %s (built %s)\n", "Cloud Resource Counter", version, date)
		os.Exit(0)
	}

	// Check whether a response file is being specified
	if cls.outputFileName != "" {
		// Try to open the file for writing
		cls.outputFile = OpenFileForWriting(cls.outputFileName, "CSV", am)
	}

	// Check whether a trace file is being specified
	if cls.traceFileName != "" {
		// Try to open the file for writing
		cls.traceFile = OpenFileForWriting(cls.traceFileName, "trace", am)
	}
}

// Display constructs a listing of all command line settings to the Activity Monitor
func (cls *CommandLineSettings) Display(resolvedRegionName string, am ActivityMonitor) {
	// What is the region being selected?
	var displayRegionName string
	if cls.allRegions {
		displayRegionName = "(All regions supported by this account)"
	} else {
		displayRegionName = resolvedRegionName
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
