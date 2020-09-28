/******************************************************************************
Cloud Resource Counter
File: main.go

Summary: Top-level entry point for the tool. Provides main() function.
******************************************************************************/

package main

import "os"

// The version of this tool. This is supplied by the build process.
var version string = "?.?.?"

// This is the build date of this tool. This is also supplied by the build process.
// TODO Replace this variable with a better name (e.g., "buildDate"). This is the
// default variable specified by Goreleaser's ldflags settings.
var date string = "<<never built>>"

// The cloud resource counter utility known as "cloud-resource-counter" inspects
// a cloud deployment (for now, only Amazon Web Services) to assess the number of
// distinct computing resources. The result is a CSV file that describes the counts
// of each.
//
// This command requires access to a valid AWS Account. For now, it is assumed that
// this is stored in the user's ".aws" folder (located in $HOME/.aws).
//
// A future version of this will allow the caller to supply credentials in more
// flexible ways.
//
func main() {
	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Command line processing
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Construct an object to send activity to the terminal (STDERR)
	monitor := &TerminalActivityMonitor{
		Writer: os.Stderr,
	}

	// Process all command line arguments
	ProcessCommandLine(monitor)

	// If we are writing to a trace file, remember to close it
	if traceFile != nil {
		defer traceFile.Close()
	}

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Establish a valid AWS Session
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching. See the session package for
	// more information.

	// Create an AWS Service Factory
	serviceFactory := &AWSServiceFactory{}
	serviceFactory.Init()

	// Show command line settings (passing in the resolved region)
	DisplayCommandLineSettings(*serviceFactory.Session.Config.Region, monitor)

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Collect counts of all resources
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Show activity
	monitor.Message("\nActivity\n")

	// Construct an array of results (this is how the results are ordered in the CSV)
	var resultData [2][]string

	// Append account ID to the result data
	AppendResults(&resultData, "Account ID", GetAccountID(serviceFactory.GetAccountIDService(), monitor))
	AppendResults(&resultData, "# of EC2 Instances", EC2Counts(serviceFactory.Session, monitor))
	AppendResults(&resultData, "# of Spot Instances", SpotInstances(serviceFactory.Session, monitor))
	AppendResults(&resultData, "# of RDS Instances", RDSInstances(serviceFactory.Session, monitor))
	AppendResults(&resultData, "# of S3 Buckets", S3Buckets(serviceFactory.Session, monitor))
	AppendResults(&resultData, "# of Lambda Functions", LambdaFunctions(serviceFactory.Session, monitor))

	// Blech: get a slice of the result data so that it can be used with WriteAll
	var csvData [][]string
	csvData = resultData[0:2]

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Construct CSV Output
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Save our results to a CSV file
	SaveToCSV(csvData, outputFile, monitor)

	// Indicate success
	monitor.Message("\nSuccess.\n")
}
