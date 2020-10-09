/******************************************************************************
Cloud Resource Counter
File: containers.go

Summary: Counts the number of unique container images.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	color "github.com/logrusorgru/aurora"
)

// UniqueContainerImages reviews all of the ECS containers either in the current region
// or (if allRegions is true) in all regions. It inspects the task definitions for all
// containers, looking at the image definition. It then counts the number of unique
// images across all containers in the given region (or all regions).
func UniqueContainerImages(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving Unique container counts")

	// Should we get the counts for all regions?
	var containerImageMap map[string]bool = make(map[string]bool)
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the container image names for a specific region
			containerImagesSlice := containerImagesForSingleRegion(sf.GetContainerService(regionName), am)

			// Add the container names to our map
			for _, cntrImg := range containerImagesSlice {
				containerImageMap[cntrImg] = true
			}
		}
	} else {
		// Get the container image names for a specific region
		containerImagesSlice := containerImagesForSingleRegion(sf.GetContainerService(""), am)

		// Add the container names to our map
		for _, cntrImg := range containerImagesSlice {
			containerImageMap[cntrImg] = true
		}
	}

	// Get our container count
	containerCount := len(containerImageMap)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(containerCount))

	return containerCount
}

// Get a list of all container images used by all tasks for this region
func containerImagesForSingleRegion(cs *ContainerService, am ActivityMonitor) []string {
	// Construct our input to find all Task Definitions
	input := &ecs.ListTaskDefinitionsInput{}

	// Indicate activity
	am.Message(".")

	// Invoke our service
	var containerImageNames []string
	err := cs.ListTaskDefinitions(input, func(page *ecs.ListTaskDefinitionsOutput, lastPage bool) bool {
		// Loop through the results...
		for _, taskDefnArn := range page.TaskDefinitionArns {
			// Construct an input struct for the specific task definition
			input := &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: taskDefnArn,
			}

			// Inspect the task definition details
			taskDefn, err := cs.InspectTaskDefinition(input)

			// Error?
			if am.CheckError(err) {
				// Stop iterating
				return false
			}

			// Do we have a TaskDefinition?
			if taskDefn.TaskDefinition != nil {
				// Loop through the container definitions...
				for _, cntrDefn := range taskDefn.TaskDefinition.ContainerDefinitions {
					containerImageNames = append(containerImageNames, *cntrDefn.Image)
				}
			}
		}

		return true
	})

	// Check for error
	am.CheckError(err)

	return containerImageNames
}
