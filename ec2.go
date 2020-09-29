/******************************************************************************
Cloud Resource Counter
File: ec2.go

Summary: Provides a count of all (non-spot) EC2 instances.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/ec2"

	color "github.com/logrusorgru/aurora"
)

// EC2Counts retrieves the count of all EC2 instances either for all
// regions (all is true) or the region associated with the
// session. This method gives status back to the user via the supplied
// ActivityMonitor instance.
func EC2Counts(awssf *AWSServiceFactory, am ActivityMonitor, all bool) string {
	// Indicate activity
	am.StartAction("Retrieving EC2 counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if all {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(awssf.Session, am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the EC2 counts for a specific region
			instanceCount += ec2CountForSingleRegion(awssf.GetEC2InstanceService(regionName), am)
		}
	} else {
		// Get the EC2 counts for the region selected by this session
		instanceCount = ec2CountForSingleRegion(awssf.GetEC2InstanceService(""), am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return strconv.Itoa(instanceCount)
}

// Get the EC2 Instance count for a single region
func ec2CountForSingleRegion(ec2s *EC2InstanceService, am ActivityMonitor) int {
	// Indicate activity
	am.Message(".")

	// Construct our input to find all EC2 instances
	input := &ec2.DescribeInstancesInput{}

	// Invoke our service
	instanceCount := 0
	err := ec2s.InspectInstances(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		// Loop through each reservation, instance
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				// Is this a valid instance? Spot instances have an InstanceLifecycle of "spot".
				// Similarly, Scheduled instances have an InstanceLifecycle of "scheduled".
				if instance.InstanceLifecycle == nil {
					instanceCount++
				}
			}
		}

		return !lastPage
	})

	// Check for error
	am.CheckError(err)

	return instanceCount
}
