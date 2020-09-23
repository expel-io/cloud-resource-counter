/******************************************************************************
Cloud Resource Counter
File: rds.go

Summary: Provides a count of all RDS instances.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	color "github.com/logrusorgru/aurora"
)

// RDSInstances does stuff...
func RDSInstances(sess *session.Session) string {
	// Create a new instance of the RDS service using the session supplied
	svc := rds.New(sess)

	// Construct our input to find all RDS instances
	input := &rds.DescribeDBInstancesInput{}

	// Indicate activity
	DisplayActivity(" * Retrieving RDS instance counts...")

	// Invoke our service
	instanceCount := 0
	err := svc.DescribeDBInstancesPages(input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		instanceCount += len(page.DBInstances)

		return !lastPage
	})

	// Check for error
	InspectError(err)

	// Indicate end of activity
	DisplayActivity("OK (%d)\n", color.Bold(instanceCount))

	return strconv.Itoa(instanceCount)
}
