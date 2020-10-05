/******************************************************************************
Cloud Resource Counter
File: lambda.go

Summary: Provides a count of all Lambda functions.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/service/lambda"

	color "github.com/logrusorgru/aurora"
)

// LambdaFunctions retrieves the count of all lambda function
// either for all regions (allRegions is true) or the region
// associated with the session.  This method gives status back
// to the user via the supplied ActivityMonitor instance.
func LambdaFunctions(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving Lambda function counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the Lambda counts for a specific region
			instanceCount += lambdaFunctionsForSingleRegion(sf.GetLambdaService(regionName), am)
		}
	} else {
		// Get the Lambda counts for the region selected by this session
		instanceCount = lambdaFunctionsForSingleRegion(sf.GetLambdaService(""), am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

func lambdaFunctionsForSingleRegion(ls *LambdaService, am ActivityMonitor) int {
	// Construct our input to find all Lambda instances
	input := &lambda.ListFunctionsInput{}

	// Indicate activity
	am.Message(".")

	// Invoke our service
	functionCounts := 0
	err := ls.ListFunctions(input, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		functionCounts += len(page.Functions)

		return true
	})

	// Check for error
	am.CheckError(err)

	return functionCounts
}
