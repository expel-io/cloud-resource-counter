/******************************************************************************
Cloud Resource Counter
File: s3.go

Summary: Provides a count of all S3 buckets.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	color "github.com/logrusorgru/aurora"
)

// S3Buckets does stuff...
func S3Buckets(sess *session.Session) string {
	// Create a new instance of the S3 service using the session supplied
	svc := s3.New(sess)

	// Construct our input to find all RDS instances
	input := &s3.ListBucketsInput{}

	// Indicate activity
	DisplayActivity(" * Retrieving S3 bucket counts...")

	// Invoke our service
	result, err := svc.ListBuckets(input)

	// Check for error
	InspectError(err)

	// Get our count of buckets
	count := len(result.Buckets)

	// Indicate end of activity
	DisplayActivity("OK (%d)\n", color.Bold(count))

	return strconv.Itoa(count)
}
