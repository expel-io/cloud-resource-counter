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
	allRegions bool
	regionName string

	// Output (CSV) file
	outputFileName string
	outputFile     *os.File
	appendToOutput bool
	noOutputFile   bool

	// Trace file
	traceFileName string
	traceFile     *os.File
}

// Process inspects the command line for valid arguments.
//
// Usage of cloud-resource-counter
//   --output-file OF: Write the results to file OF. Defaults to 'resources.csv'
//   --no-output:      If set, then the results are not saved to any file.
//   --profile PN:     Use the credentials associated with shared profile PN
//   --region RN:      View resource counts for the AWS region RN
//   --trace-file TF:  Create a trace file that contains all calls to AWS.
//   --version:        Display version information
//
func (cls *CommandLineSettings) Process(args []string, am ActivityMonitor) func() {
	var showVersion bool
	emptyFn := func() {}

	// What is our default profile?
	if cls.defaultProfileName = os.Getenv("AWS_PROFILE"); cls.defaultProfileName == "" {
		cls.defaultProfileName = session.DefaultSharedConfigProfile
	}

	// Define a new FlagSet
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Define and parse the command line arguments...
	flagSet.StringVar(&cls.outputFileName, "output-file", "", "CSV Output File. Specify a path to a `file` to save the generated CSV file. (default resources.csv)")
	flagSet.BoolVar(&cls.noOutputFile, "no-output", false, "Do not save the results of this run into any file. (default false--save results to a file)")
	flagSet.StringVar(&cls.profileName, "profile", cls.defaultProfileName, "The name of the AWS Profile to use.")
	flagSet.StringVar(&cls.regionName, "region", "", "The name of the AWS Region to use. If omitted, then all regions will be examined. This is the default behavior.")
	flagSet.StringVar(&cls.traceFileName, "trace-file", "", "AWS Trace Log. Specify a `file` to record API calls being made. Each subsequent run OVERWRITES the prior run.")
	flagSet.BoolVar(&showVersion, "version", false, "Shows the version number.")
	flagSet.Parse(args)

	// Check for a valid AWS Region
	if cls.regionName != "" {
		// If not valid region name, then get out now...
		if !IsValidRegionName(cls.regionName) {
			am.ActionError("Error: '%s' is not a valid AWS Region name.", cls.regionName)
			return emptyFn
		}
	} else {
		// Record that all regions are being examined
		cls.allRegions = true
	}

	// If both --output-file and --no-output specified, then complain
	if cls.outputFileName != "" && cls.noOutputFile {
		// Show error...
		am.ActionError("Error: Cannot specify both --output-file and -no-output!")
		return emptyFn
	}

	// If no output file specified, then use a default name (assuming that we are not barring output)
	if cls.outputFileName == "" && !cls.noOutputFile {
		// Set the default output file
		cls.outputFileName = "resources.csv"
	}

	// Did the user just want to see the version?
	if showVersion {
		am.Message("%s, version %s (built %s)\n", "Cloud Resource Counter", version, date)
		am.Exit(0)

		return emptyFn
	}

	// Check whether a response file is being specified
	if cls.outputFileName != "" && !cls.noOutputFile {
		// Determine whether to append the output file or not
		cls.appendToOutput = FileExists(cls.outputFileName)

		// Try to open the file for writing
		cls.outputFile = OpenFileForWriting(cls.outputFileName, "CSV", am, cls.appendToOutput)
	}

	// Check whether a trace file is being specified
	if cls.traceFileName != "" {
		// Try to open the file for writing
		cls.traceFile = OpenFileForWriting(cls.traceFileName, "trace", am, false)
	}

	// Return a function that we can use for cleaning up open resources
	return func() {
		if !NilInterface(cls.outputFile) {
			cls.outputFile.Close()
		}
		if !NilInterface(cls.traceFile) {
			cls.traceFile.Close()
		}
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

	// What is the file name for the output file
	var displayOutputFile string
	if cls.outputFileName == "" {
		displayOutputFile = "(none)"
	} else {
		displayOutputFile = cls.outputFileName
	}

	// Output information about utility running
	am.Message("%s (v%s) running with:\n", color.Bold("Cloud Resource Counter"), version)
	am.Message(" o %s: %s\n", color.Italic("AWS Profile"), cls.profileName)
	am.Message(" o %s:  %s\n", color.Italic("AWS Region"), displayRegionName)
	am.Message(" o %s: %s\n", color.Italic("Output file"), displayOutputFile)

	// Are we tracing?
	if cls.traceFileName != "" {
		am.Message(" o %s:  %s\n", color.Italic("Trace file"), cls.traceFileName)
	}
}
