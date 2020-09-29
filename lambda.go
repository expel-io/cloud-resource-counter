/******************************************************************************
Cloud Resource Counter
File: lambda.go

Summary: Provides a count of all Lambda functions.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	color "github.com/logrusorgru/aurora"
)

// LambdaFunctions retrieves the count of all lambda function.
// TODO ... either for all regions (allRegions is true) or the
// TODO ... region associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func LambdaFunctions(sess *session.Session, am ActivityMonitor) int {
	// Create a new instance of the Lambda service using the session supplied
	svc := lambda.New(sess)

	// Construct our input to find all Lambda instances
	input := &lambda.ListFunctionsInput{}

	// Indicate activity
	am.StartAction("Retrieving Lambda function counts")

	// Invoke our service
	functionCounts := 0
	err := svc.ListFunctionsPages(input, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		functionCounts += len(page.Functions)

		return !lastPage
	})

	// Check for error
	am.CheckError(err)

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(functionCounts))

	return functionCounts
}
