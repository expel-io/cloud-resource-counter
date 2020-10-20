/******************************************************************************
Cloud Resource Counter
File: ec2.go

Summary: Provides a count of all (non-spot) EC2 instances.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	color "github.com/logrusorgru/aurora"
)

// EC2Counts retrieves the count of all EC2 instances either for all
// regions (allRegions is true) or the region associated with the
// session. This method gives status back to the user via the supplied
// ActivityMonitor instance.
func EC2Counts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving EC2 counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the EC2 counts for a specific region
			instanceCount += ec2CountForSingleRegion(sf.GetEC2InstanceService(regionName), am)
		}
	} else {
		// Get the EC2 counts for the region selected by this session
		instanceCount = ec2CountForSingleRegion(sf.GetEC2InstanceService(""), am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

// Get the EC2 Instance count for a single region
func ec2CountForSingleRegion(ec2is *EC2InstanceService, am ActivityMonitor) int {
	// Indicate activity
	am.Message(".")

	// Construct our input to find only RUNNING EC2 instances
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	}

	// Invoke our service
	instanceCount := 0
	err := ec2is.InspectInstances(input, func(dio *ec2.DescribeInstancesOutput, lastPage bool) bool {
		// Loop through each reservation, instance
		for _, reservation := range dio.Reservations {
			for _, instance := range reservation.Instances {
				// Is this a valid instance? Spot instances have an InstanceLifecycle of "spot".
				// Similarly, Scheduled instances have an InstanceLifecycle of "scheduled".
				if instance.InstanceLifecycle == nil {
					instanceCount++
				}
			}
		}

		return true
	})

	// Check for error
	am.CheckError(err)

	return instanceCount
}
