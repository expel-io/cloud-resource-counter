/******************************************************************************
Cloud Resource Counter
File: ec2.go

Summary: Provides a count of all (non-spot) EC2 instances.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	color "github.com/logrusorgru/aurora"
)

// EC2Counts does stuff...
func EC2Counts(sess *session.Session) string {
	// Create a new instance of the EC2 service using the session supplied
	svc := ec2.New(sess)

	// Construct our input to find all EC2 instances
	input := &ec2.DescribeInstancesInput{}

	// Indicate activity
	DisplayActivity(" * Retrieving EC2 counts...")

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
	InspectError(err)

	// Indicate end of activity
	DisplayActivity("OK (%d)\n", color.Bold(instanceCount))

	return strconv.Itoa(instanceCount)
}
