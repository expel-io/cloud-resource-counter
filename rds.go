/******************************************************************************
Cloud Resource Counter
File: rds.go

Summary: Provides a count of all RDS instances.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"

	color "github.com/logrusorgru/aurora"
)

// RDSInstances retrieves the count of all RDS Instances.
// TODO ... either for all regions (allRegions is true) or the
// TODO ... region associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func RDSInstances(sf ServiceFactory, sess *session.Session, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving RDS instance counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the RDS instance counts for a specific region
			instanceCount += rdsInstancesForSingleRegion(sess, regionName, am)
		}
	} else {
		// Get the RDS instance counts for the region selected by this session
		instanceCount = rdsInstancesForSingleRegion(sess, "", am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

func rdsInstancesForSingleRegion(sess *session.Session, regionName string, am ActivityMonitor) int {
	// Construct our service client
	var svc *rds.RDS
	if regionName == "" {
		svc = rds.New(sess)
	} else {
		svc = rds.New(sess, aws.NewConfig().WithRegion(regionName))
	}

	// Construct our input to find all RDS instances
	input := &rds.DescribeDBInstancesInput{}

	// Indicate activity
	am.Message(".")

	// Invoke our service
	instanceCount := 0
	err := svc.DescribeDBInstancesPages(input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		instanceCount += len(page.DBInstances)

		return !lastPage
	})

	// Check for error
	am.CheckError(err)

	return instanceCount
}
