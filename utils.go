/******************************************************************************
Cloud Resource Counter
File: utils.go

Summary: Various utility functions
******************************************************************************/

package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
	color "github.com/logrusorgru/aurora"
)

// DisplayActivity displays text to the user to indicate activity
func DisplayActivity(format string, v ...interface{}) {
	fmt.Fprint(os.Stderr, fmt.Sprintf(format, v...))
}

// DisplayActivityError display error
func DisplayActivityError(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, color.Red(fmt.Sprintf(format, v...)))
	fmt.Fprintln(os.Stderr)

	os.Exit(1)
}

// InspectError inspects the supplied error and reports accordingly
func InspectError(err error) {
	// If it is nil, get out now!
	if err == nil {
		return
	}

	// Is this an AWS Error?
	if aerr, ok := err.(awserr.Error); ok {
		// Switch on the error code for known error conditions...
		switch aerr.Code() {
		case "NoCredentialProviders":
			// TODO Can we establish this failure earlier? When the session is created?
			DisplayActivityError("Either the profile name is misspelled or credentials are not stored.")
			break
		default:
			DisplayActivityError("%v", aerr)
		}
	} else {
		DisplayActivityError("%v", err)
	}
}

// AppendResults is used to grow our results data structure
func AppendResults(results *[2][]string, colName string, colValue string) {
	results[0] = append(results[0], colName)
	results[1] = append(results[1], colValue)
}

// SaveToCSV saves the data structure to a CSV file
func SaveToCSV(csvData [][]string, outputFileName string) {
	// Indicate activity
	DisplayActivity(" * Writing to file...")

	// Open the file
	recordFile, err := os.Create(outputFileName)

	// Check for error
	InspectError(err)

	// Remember to close the file...
	defer recordFile.Close()

	// Get the CSV Writer
	writer := csv.NewWriter(recordFile)

	// Write all of the contents at once
	err = writer.WriteAll(csvData)

	// Check for Error
	InspectError(err)

	// Indicate success
	DisplayActivity("OK\n")
}
