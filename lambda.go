/******************************************************************************
Cloud Resource Counter
File: lambda.go

Summary: Provides a count of all Lambda functions.
******************************************************************************/

package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	color "github.com/logrusorgru/aurora"
)

// LambdaFunctions does stuff...
func LambdaFunctions(sess *session.Session) string {
	// Create a new instance of the Lambda service using the session supplied
	svc := lambda.New(sess)

	// Construct our input to find all Lambda instances
	input := &lambda.ListFunctionsInput{}

	// Indicate activity
	DisplayActivity(" * Retrieving Lambda function counts...")

	// Invoke our service
	functionCounts := 0
	err := svc.ListFunctionsPages(input, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		functionCounts += len(page.Functions)

		return !lastPage
	})

	// Check for error
	InspectError(err)

	// Indicate end of activity
	DisplayActivity("OK (%d)\n", color.Bold(functionCounts))

	return strconv.Itoa(functionCounts)
}
