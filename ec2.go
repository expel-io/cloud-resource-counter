/******************************************************************************
Cloud Resource Counter
File: ec2.go

Summary: Provides a count of all (non-spot) EC2 instances.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	color "github.com/logrusorgru/aurora"
)

// EC2Counts retrieves the count of all EC2 instances either for all
// regions (allRegions is true) or the region associated with the
// session. This method gives status back to the user via the supplied
// ActivityMonitor instance.
func EC2Counts(sess *session.Session, am ActivityMonitor) string {
	// Indicate activity
	am.StartAction("Retrieving EC2 counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sess, am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the EC2 counts for a specific region
			instanceCount += ec2CountForSingleRegion(sess, regionName, am)
		}
	} else {
		// Get the EC2 counts for the region selected by this session
		instanceCount = ec2CountForSingleRegion(sess, "", am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return strconv.Itoa(instanceCount)
}

// Get the EC2 Instance count for a single region
func ec2CountForSingleRegion(sess *session.Session, regionName string, am ActivityMonitor) int {
	// Indicate activity
	am.Message(".")

	// Determine the session to use for this call
	var svcSession *session.Session
	if regionName == "" {
		svcSession = sess
	} else {
		svcSession = sess.Copy(&aws.Config{Region: aws.String(regionName)})
	}

	// Create a new instance of the EC2 service using the session supplied
	svc := ec2.New(svcSession)

	// Construct our input to find all EC2 instances
	input := &ec2.DescribeInstancesInput{}

	// Invoke our service
	instanceCount := 0
	err := svc.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
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
