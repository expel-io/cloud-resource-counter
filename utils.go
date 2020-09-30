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
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AppendResults is used to grow our results data structure
func AppendResults(results *[2][]string, colName string, colValue interface{}) {
	results[0] = append(results[0], colName)
	results[1] = append(results[1], fmt.Sprintf("%v", colValue))
}

// SaveToCSV saves the data structure to a CSV file
func SaveToCSV(csvData [][]string, file *os.File, am ActivityMonitor) {
	// Indicate activity
	am.StartAction("Writing to file")

	// Remember to close the file...
	defer file.Close()

	// Get the CSV Writer
	writer := csv.NewWriter(file)

	// Write all of the contents at once
	err := writer.WriteAll(csvData)

	// Check for Error
	am.CheckError(err)

	// Indicate success
	am.EndAction("OK")
}

// OpenFileForWriting does stuff...
func OpenFileForWriting(fileName string, typeOfFile string, am ActivityMonitor) *os.File {
	// Can we open it for writing?
	file, err := os.Create(fileName)

	// Check for error
	am.CheckError(err)

	return file
}

// GetEC2Regions determines the set of regions associated with the account.
func GetEC2Regions(ec2is *EC2InstanceService, am ActivityMonitor) []string {
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
	result, err := ec2is.GetRegions(input)

	// Do we have an error?
	if am.CheckError(err) {
		return nil
	}

	// Transform the array of results into an array of region names...
	var regionNames []string
	for _, regionInfo := range result.Regions {
		regionNames = append(regionNames, *regionInfo.RegionName)
	}

	return regionNames
}

// Map applies a function to each element of a string array
// Borrowed from: https://gobyexample.com/collection-functions
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
