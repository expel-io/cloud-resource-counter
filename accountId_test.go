package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/expel-io/cloud-resource-counter/mock"
)

type mockedGetCallerIdentity struct {
	stsiface.STSAPI
	Resp *sts.GetCallerIdentityOutput
}

func (m *mockedGetCallerIdentity) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if m.Resp != nil {
		return m.Resp, nil
	}

	return nil, fmt.Errorf("Unable to generate an account ID--you don't have a valid session")
}

func TestGetAccountID(t *testing.T) {
	// Create a couple of cases
	cases := []struct {
		Resp        *sts.GetCallerIdentityOutput
		AccountID   string
		ExpectError bool
	}{
		{
			&sts.GetCallerIdentityOutput{
				Account: aws.String("123abc"),
			},
			"123abc",
			false,
		}, {
			nil,
			"",
			true,
		},
	}

	for _, c := range cases {
		// Create a new services...
		svc := &CallerIdentityService{
			Client: &mockedGetCallerIdentity{
				Resp: c.Resp,
			},
		}

		// Create a mock activity monitor
		mon := &mock.MockedActivityMonitor{}

		// Get the account ID
		accountID := GetAccountID(svc, mon)

		// Do we expect?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			t.Error("Did not expect an error, but it occured!")
		} else if accountID != c.AccountID {
			t.Errorf("Account returned '%s'; expected %s", accountID, c.AccountID)
		}
	}
}
