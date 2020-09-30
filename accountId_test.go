package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/expel-io/cloud-resource-counter/mock"
)

// This type "stands in" for the real SecurityTokenService. It implements
// the GetCallerIdentity method by allowing the caller to indicate the
// desired GetCallerIdentityOutput struct.
type fakeSecurityTokenService struct {
	stsiface.STSAPI
	Resp *sts.GetCallerIdentityOutput
}

// This fake GetCallerIdentity method takes an arbitrary input and returns
// either an error (if the supplied response object is nil) or the supplied
// response (in the form of a GetCallerIdentityOutput pointer).
func (m *fakeSecurityTokenService) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	// Was the provided Response present?
	if m.Resp != nil {
		// Return it with no error
		return m.Resp, nil
	}

	// Return an error
	return nil, fmt.Errorf("Unable to generate an account ID--you don't have a valid session")
}

// This function tests two scenarios: GetCallerIdentity succeeds and fails.
func TestGetAccountID(t *testing.T) {
	// Fake Account ID
	const fakeAccountID = "123abc"

	// Create two test cases
	cases := []struct {
		Resp              *sts.GetCallerIdentityOutput
		expectedAccountID string
		ExpectError       bool
	}{
		{
			&sts.GetCallerIdentityOutput{
				Account: aws.String(fakeAccountID),
			},
			fakeAccountID,
			false,
		}, {
			nil,
			"",
			true,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create a AccountIDService with a fake client
		svc := &AccountIDService{
			Client: &fakeSecurityTokenService{
				Resp: c.Resp,
			},
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Get the account ID
		actualAccountID := GetAccountID(svc, mon)

		// Do we expect an error to occur?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			// Did an error occur?
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else if actualAccountID != c.expectedAccountID {
			// Does the expected value NOT match the actual value?
			t.Errorf("Error: Account returned '%s'; expected %s", actualAccountID, c.expectedAccountID)
		}
	}
}
