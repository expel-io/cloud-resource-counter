/******************************************************************************
Cloud Resource Counter
File: accountId.go

Summary: Retrieve account ID (assumed to be a single value) for the current
         user session.
******************************************************************************/

package main

import (
	color "github.com/logrusorgru/aurora"
)

// GetAccountID returns the Amazon Account ID for the supplied session.
func GetAccountID(cis *CallerIdentityService, am ActivityMonitor) string {
	// Indicate activity
	am.StartAction("Retrieving Account ID")

	// Get the caller's identity
	accountID, err := cis.Account()

	// Check for error
	am.CheckError(err)

	// Indicate end of activity
	am.EndAction("OK (%s)", color.Bold(accountID))

	return accountID
}
