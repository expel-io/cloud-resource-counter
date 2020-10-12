/******************************************************************************
Cloud Resource Counter
File: ebs.go

Summary: Count the number of EBS Volumes
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	color "github.com/logrusorgru/aurora"
)

// EBSVolumes returns a count of all EBS volumes in the current region (if allRegions
// is false) or in all regions associated with this account (if allRegions is true).
func EBSVolumes(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving EBS volume counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the EBS Volume counts for a specific region
			instanceCount += ebsVolumesForSingleRegion(sf.GetEC2InstanceService(regionName), am)
		}
	} else {
		// Get the EBS Volume counts for the region selected by this session
		instanceCount = ebsVolumesForSingleRegion(sf.GetEC2InstanceService(""), am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

func ebsVolumesForSingleRegion(ec2is *EC2InstanceService, am ActivityMonitor) int {
	// Indicate activity
	am.Message(".")

	// Construct our input to find all EBS volumes
	input := &ec2.DescribeVolumesInput{}

	// Invoke our service
	instanceCount := 0
	err := ec2is.InspectVolumes(input, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		// Loop through each Volume
		for _, volume := range page.Volumes {
			// Do we have a non-nil, non-empty Attachments array?
			if volume.Attachments != nil && len(volume.Attachments) > 0 {
				instanceCount++
			}
		}

		return true
	})

	// Check for error
	am.CheckError(err)

	return instanceCount
}
