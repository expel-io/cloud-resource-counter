/******************************************************************************
Cloud Resource Counter
File: s3.go

Summary: Provides a count of all S3 buckets.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/service/s3"

	color "github.com/logrusorgru/aurora"
)

// S3Buckets retrieves the count of all S3 buckets in ALL REGIONS.
// This behavior is unlike other AWS Services (e.g., EC2, Spot, RDS,
// etc).
//
// AS SUCH, THIS COUNT WILL BE INCORRECT WHEN A SINGLE REGION IS SPECIFIED.
//
// It is unclear if displaying counts of distinct regions will be part
// of the final CLI.
//
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func S3Buckets(sf ServiceFactory, am ActivityMonitor) int {
	// Create a new instance of the S3 (abstract) service
	svc := sf.GetS3Service()

	// Construct our input to find all S3 buckets
	input := &s3.ListBucketsInput{}

	// Indicate activity
	am.StartAction("Retrieving S3 bucket counts")

	// Invoke our service
	result, err := svc.ListBuckets(input)

	// Check for error
	if am.CheckError(err) {
		return 0
	}

	// Get our count of buckets
	count := len(result.Buckets)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(count))

	return count
}
