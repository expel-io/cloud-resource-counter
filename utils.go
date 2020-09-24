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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
func SaveToCSV(csvData [][]string, file *os.File) {
	// Indicate activity
	DisplayActivity(" * Writing to file...")

	// Remember to close the file...
	defer file.Close()

	// Get the CSV Writer
	writer := csv.NewWriter(file)

	// Write all of the contents at once
	err := writer.WriteAll(csvData)

	// Check for Error
	InspectError(err)

	// Indicate success
	DisplayActivity("OK\n")
}

// OpenFileForWriting does stuff...
func OpenFileForWriting(fileName string, typeOfFile string) *os.File {
	// Can we open it for writing?
	file, err := os.Create(fileName)

	// Check for error
	if err != nil {
		// Construct an error message
		message := color.Red(fmt.Sprintf("Unable to open %s file for writing => %v", typeOfFile, err))

		// Display error
		fmt.Fprintln(os.Stderr, message)

		os.Exit(1)
	}

	return file
}

// GetEC2Regions does stuff...
func GetEC2Regions(sess *session.Session) []string {
	// Get a new instances of the EC2 service
	svc := ec2.New(sess)

	// Construct the input
	input := &ec2.DescribeRegionsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("opt-in-status"),
				Values: []*string{
					aws.String("opt-in-not-required"),
					aws.String("opted-in"),
				},
			},
		},
	}

	// Execute the command
	result, err := svc.DescribeRegions(input)

	// Do we have an error?
	if err != nil {
		// Display a message and then let's exit
		fmt.Fprintln(os.Stderr, color.Red(fmt.Sprintf("Unable to get a list of valid EC2 regions: %v", err)))

		os.Exit(1)
	}

	// Transform the array of results into an array of region names...
	var regionNames []string
	for _, regionInfo := range result.Regions {
		regionNames = append(regionNames, *regionInfo.RegionName)
	}

	return regionNames
}
