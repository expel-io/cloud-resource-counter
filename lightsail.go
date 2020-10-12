/******************************************************************************
Cloud Resource Counter
File: lightsail.go

Summary: Counts the number of Lightsail instances.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/service/lightsail"
	color "github.com/logrusorgru/aurora"
)

// LightsailInstances returns a count of Lightsail instances in the current region
// (allRegions = false) or for all regions (allRegions = true)
func LightsailInstances(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving Lightsail instance counts")

	// Input for the list of regions...
	input := &lightsail.GetRegionsInput{}

	// Get the list of all enabled regions for this account
	// Note that this call fails if the default region associated with this
	// account is not in the supported list. Must use something supported,
	// like US-EAST-1.
	response, err := sf.GetLightsailService(DefaultRegion).GetRegions(input)

	// If error, then get out now!
	if am.CheckError(err) {
		return 0
	}

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Loop through all of the regions
		for _, region := range response.Regions {
			// Get the Lightsail instances counts for a specific region
			instanceCount += lightsailInstancesForSingleRegion(sf.GetLightsailService(*region.Name), am)
		}
	} else {
		// Is the current region supported by Lightsail?
		var validLightsailRegion bool
		for _, region := range response.Regions {
			if sf.GetCurrentRegion() == *region.Name {
				validLightsailRegion = true
			}
		}

		if validLightsailRegion {
			// Get the Lightsail instances counts for the region selected by this session
			instanceCount = lightsailInstancesForSingleRegion(sf.GetLightsailService(""), am)
		}
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

func lightsailInstancesForSingleRegion(lss *LightsailService, am ActivityMonitor) int {
	// Construct our input to find all Lightsail instances
	input := &lightsail.GetInstancesInput{}

	// Indicate activity
	am.Message(".")

	// Invoke our service
	response, err := lss.InspectInstances(input)

	// Check for error
	if am.CheckError(err) {
		return 0
	}

	return len(response.Instances)
}
