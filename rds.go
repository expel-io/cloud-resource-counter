/******************************************************************************
Cloud Resource Counter
File: rds.go

Summary: Provides a count of all RDS instances.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"

	color "github.com/logrusorgru/aurora"
)

// RDSInstances retrieves the count of all RDS Instances.
// TODO ... either for all regions (allRegions is true) or the
// TODO ... region associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func RDSInstances(sess *session.Session, am ActivityMonitor) int {
	// Create a new instance of the RDS service using the session supplied
	svc := rds.New(sess)

	// Construct our input to find all RDS instances
	input := &rds.DescribeDBInstancesInput{}

	// Indicate activity
	am.StartAction("Retrieving RDS instance counts")

	// Invoke our service
	instanceCount := 0
	err := svc.DescribeDBInstancesPages(input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		instanceCount += len(page.DBInstances)

		return !lastPage
	})

	// Check for error
	am.CheckError(err)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}
