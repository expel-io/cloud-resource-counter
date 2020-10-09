/******************************************************************************
Cloud Resource Counter
File: utils.go

Summary: Various utility functions
******************************************************************************/

package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// OpenFileForWriting does stuff...
func OpenFileForWriting(fileName string, typeOfFile string, am ActivityMonitor, append bool) *os.File {
	// What is our flag for the file?
	var flag int
	if append {
		flag = os.O_WRONLY | os.O_APPEND
	} else {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	// Can we open it for writing?
	file, err := os.OpenFile(fileName, flag, 0666)

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
