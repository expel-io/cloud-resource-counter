/******************************************************************************
Cloud Resource Counter
File: accountId.go

Summary: Retrieve account ID (assumed to be a single value) for the current
         user session.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// GetAccountID returns the Amazon Account ID for the supplied session.
func GetAccountID(sess *session.Session) string {
	// Create a new instance of the Security Token Service's client with a Session.
	svc := sts.New(sess)

	// Indicate activity
	DisplayActivity(" * Retrieving Account ID...")

	// Construct the input parameter
	input := &sts.GetCallerIdentityInput{}

	// Get the caller's identity
	result, err := svc.GetCallerIdentity(input)

	// Check for error
	InspectError(err)

	// Indicate end of activity
	DisplayActivity("OK\n")

	return *result.Account
}
