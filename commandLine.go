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

	color "github.com/logrusorgru/aurora"
)

var profileName, outputFileName, regionName string

// ProcessCommandLine inspects the command line for valid arguments
//
// Usage of cloud-resource-counter
//   --output-file OF: Write the results to file OF
//   --profile PN:     Use the credentials associated with shared profile PN
//   --region RN:      View resource counts for the AWS region RN
//   --version:        Display version information
//
func ProcessCommandLine() {
	var showVersion bool

	// What is our default profile?
	var defaultProfileName string
	if defaultProfileName = os.Getenv("AWS_PROFILE"); defaultProfileName == "" {
		defaultProfileName = "default"
	}

	// What is our default region
	defaultRegionName := os.Getenv("AWS_REGION")

	// Define and parse the command line arguments...
	flag.StringVar(&profileName, "profile", defaultProfileName, "AWS Profile Name")
	flag.StringVar(&outputFileName, "output-file", "./resources.csv", "CSV Output File")
	flag.StringVar(&regionName, "region", defaultRegionName, "Selects an AWS Region to use")
	flag.BoolVar(&showVersion, "version", false, "Shows the version number.")
	flag.Parse()

	// TODO Check for a valid AWS Region

	// TODO Check for missing required arguments. If any, show usage and quit

	// Did the user just want to see the version?
	if showVersion {
		DisplayActivity("%s, version %s\n", "Cloud Resource Counter", version)
		os.Exit(0)
	}

	// What is the region being selected?
	var displayRegionName string
	if displayRegionName = regionName; displayRegionName == "" {
		displayRegionName = "(specified by the profile)"
	}

	// Output information about utility running
	DisplayActivity("%s (v%s) running with:\n", color.Bold("Cloud Resource Counter"), version)
	DisplayActivity(" o %s: %s\n", color.Italic("AWS Profile"), profileName)
	DisplayActivity(" o %s:  %s\n", color.Italic("AWS Region"), displayRegionName)
	DisplayActivity(" o %s: %s\n", color.Italic("Output file"), outputFileName)
}
