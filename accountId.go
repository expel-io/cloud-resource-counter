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

// GetAccountID returns the Amazon Account ID for the supplied session, showing activity
// in the process and handling potential errors.
// It relies a supplied AccountIDService struct which has a single method: Account.
// It also relies on a supplied ActivityMonitor which it uses to inform the user of
// what it is doing.
func GetAccountID(cis *AccountIDService, am ActivityMonitor) string {
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
