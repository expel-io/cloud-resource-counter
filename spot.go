/******************************************************************************
Cloud Resource Counter
File: spot.go

Summary: Provides a count of all Spot EC2 instances.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	color "github.com/logrusorgru/aurora"
)

// SpotInstances retrieves the count of all EC2 spot instances.
// TODO ... either for all regions (allRegions is true) or the
// TODO ... region associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func SpotInstances(sess *session.Session, am ActivityMonitor) string {
	// Create a new instance of the EC2 service using the session supplied
	svc := ec2.New(sess)

	// Construct our input to find all EC2 instances
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-lifecycle"),
				Values: []*string{
					aws.String("spot"),
				},
			},
		},
	}

	// Indicate activity
	am.StartAction("Retrieving Spot instance counts")

	// Invoke our service
	instanceCount := 0
	err := svc.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		// Loop through each reservation
		for _, reservation := range page.Reservations {
			instanceCount += len(reservation.Instances)
		}

		return !lastPage
	})

	// Check for error
	am.CheckError(err)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return strconv.Itoa(instanceCount)
}
