/******************************************************************************
Cloud Resource Counter
File: s3.go

Summary: Provides a count of all S3 buckets.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	color "github.com/logrusorgru/aurora"
)

// S3Buckets retrieves the count of all S3 buckets.
// TODO ... either for all regions (allRegions is true) or the
// TODO ... region associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func S3Buckets(sess *session.Session, am ActivityMonitor) int {
	// Create a new instance of the S3 service using the session supplied
	svc := s3.New(sess)

	// Construct our input to find all RDS instances
	input := &s3.ListBucketsInput{}

	// Indicate activity
	am.StartAction("Retrieving S3 bucket counts")

	// Invoke our service
	result, err := svc.ListBuckets(input)

	// Check for error
	am.CheckError(err)

	// Get our count of buckets
	count := len(result.Buckets)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(count))

	return count
}
