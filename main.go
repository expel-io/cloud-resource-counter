/******************************************************************************
Cloud Resource Counter
File: main.go

Summary: Top-level entry point for the tool. Provides main() function.
******************************************************************************/

package main

// The version of this tool. This needs to be exported into a file that can be
// loaded into this program (and created by the build process).
const version = "0.2"

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

	// Process all command line arguments
	ProcessCommandLine()

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
	sess := EstablishAwsSession()

	// Show command line settings
	DisplayCommandLineSettings(*sess.Config.Region)

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Collect counts of all resources
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Show activity
	DisplayActivity("\nActivity\n")

	// Construct an array of results (this is how the results are ordered in the CSV)
	var resultData [2][]string

	// Append account ID to the result data
	AppendResults(&resultData, "Account ID", GetAccountID(sess))
	AppendResults(&resultData, "# of EC2 Instances", EC2Counts(sess))
	AppendResults(&resultData, "# of Spot Instances", SpotInstances(sess))
	AppendResults(&resultData, "# of RDS Instances", RDSInstances(sess))
	AppendResults(&resultData, "# of S3 Buckets", S3Buckets(sess))
	AppendResults(&resultData, "# of Lambda Functions", LambdaFunctions(sess))

	// Blech: get a slice of the result data so that it can be used with WriteAll
	var csvData [][]string
	csvData = resultData[0:2]

	/* =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
	 * Construct CSV Output
	 * =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-= */

	// Save our results to a CSV file
	SaveToCSV(csvData, outputFile)

	// Show activity
	DisplayActivity("\nSuccess.\n")
}
