/******************************************************************************
Cloud Resource Counter
File: main.go

Summary: Top-level entry point for the tool. Provides main() function.
******************************************************************************/

package main

import (
	"os"
	"time"
)

// The version of this tool. This is supplied by the build process.
// This is default variable used by Goreleaser's ldflags settings. If
// we want to use a different variable, we need to change to Goreleaser
// settings file (which is .goreleaser.yml).
var version string = "?.?.?"

// This is the build date of this tool. This is also supplied by the build process.
// TODO Replace this variable with a better name (e.g., "buildDate"). This is the
// default variable specified by Goreleaser's ldflags settings.
var date string = "<<never built>>"

// The cloud resource counter utility known as "aws-resource-counter" inspects
// a cloud deployment (for now, only Amazon Web Services) to assess the number of
// distinct computing resources. The result is a CSV file that describes the counts
// of each.
//
// This command requires access to a valid AWS Account. For now, it is assumed that
// this is stored in the user's ".aws" folder (located in $HOME/.aws).
//
func main() {
	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Command line processing
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Construct an object to be our Activity Monitor. This will send results to
	// the Terminal (Standard Error).
	monitor := &TerminalActivityMonitor{
		Writer: os.Stderr,
		ExitFn: os.Exit,
	}

	// Process all command line arguments
	settings := &CommandLineSettings{}
	cleanupFn := settings.Process(os.Args[1:], monitor)
	defer cleanupFn()

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Establish a valid AWS Session via our AWS Service Factory
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Create an AWS Service Factory
	serviceFactory := &AWSServiceFactory{
		ProfileName: settings.profileName,
		RegionName:  settings.regionName,
		TraceWriter: settings.traceFile,
		UseSSO:      settings.useSSO,
	}
	serviceFactory.Init()

	// Show command line settings
	settings.Display(monitor)

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Collect counts of all resources
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Show activity
	monitor.Message("\nActivity\n")

	// Construct a new results data structure
	results := Results{
		StoreHeaders: !settings.appendToOutput,
		Writer:       settings.outputFile,
	}
	results.Init()

	// Get the display name of the selected region
	var displayRegion string
	if settings.allRegions {
		displayRegion = "ALL_REGIONS"
	} else {
		displayRegion = settings.regionName
	}

	// Create a new row of data
	results.NewRow()
	results.Append("Account ID", GetAccountID(serviceFactory.GetAccountIDService(), monitor))
	results.Append("Timestamp", time.Now().Format(time.RFC3339))
	results.Append("Region", displayRegion)
	results.Append("# of EC2 Instances", EC2Counts(serviceFactory, monitor, settings.allRegions))
	results.Append("# of Spot Instances", SpotInstances(serviceFactory, monitor, settings.allRegions))
	results.Append("# of EBS Volumes", EBSVolumes(serviceFactory, monitor, settings.allRegions))
	results.Append("# of Unique Containers", UniqueContainerImages(serviceFactory, monitor, settings.allRegions))
	results.Append("# of Lambda Functions", LambdaFunctions(serviceFactory, monitor, settings.allRegions))
	results.Append("# of RDS Instances", RDSInstances(serviceFactory, monitor, settings.allRegions))
	results.Append("# of Lightsail Instances", LightsailInstances(serviceFactory, monitor, settings.allRegions))
	results.Append("# of S3 Buckets", S3Buckets(serviceFactory, monitor, settings.allRegions))

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Construct CSV Output
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Save our results to a CSV file
	results.Save(monitor)

	// Do we need to "explain" our S3 count?
	if !settings.allRegions {
		monitor.Message("\n*S3 counts cannot be computed on a per-region basis. This count is for ALL REGIONS.\n")
	}

	// Indicate success
	monitor.Message("\nSuccess.\n")
}
