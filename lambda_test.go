/******************************************************************************
Cloud Resource Counter
File: lambda_test.go

Summary: The Unit Test for lambda.
******************************************************************************/

package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Lambda Function Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our map of regions and the functions in each
var lambdaFnsPerRegion = map[string][]*lambda.ListFunctionsOutput{
	// US-EAST-1 illustrates a case where ListFunctionsPages returns 1
	// page of 4 results
	"us-east-1": []*lambda.ListFunctionsOutput{
		&lambda.ListFunctionsOutput{
			Functions: []*lambda.FunctionConfiguration{
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
			},
		},
	},
	// US-EAST-2 illustrates a case where ListFunctionsPages returns 1 page of
	// 3 results
	"us-east-2": []*lambda.ListFunctionsOutput{
		&lambda.ListFunctionsOutput{
			Functions: []*lambda.FunctionConfiguration{
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
			},
		},
	},
	// AF-SOUTH-1 illustrates a case where ListFunctionsPages returns two pages
	// of results.
	// First page: 9 functions
	// Second page: 1 functions
	"af-south-1": []*lambda.ListFunctionsOutput{
		&lambda.ListFunctionsOutput{
			Functions: []*lambda.FunctionConfiguration{
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
				&lambda.FunctionConfiguration{},
			},
		},
		&lambda.ListFunctionsOutput{
			Functions: []*lambda.FunctionConfiguration{
				&lambda.FunctionConfiguration{},
			},
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Lambda Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// To use this struct, the caller must supply a ListFunctionsOutput slice. If
// it is missing, it will trigger the mock function to simulate an error from
// the corresponding function.
type fakeLambdaService struct {
	lambdaiface.LambdaAPI
	LFOResponse []*lambda.ListFunctionsOutput
}

// Simulate the ListFunctionsPages function
func (fake *fakeLambdaService) ListFunctionsPages(input *lambda.ListFunctionsInput, fn func(*lambda.ListFunctionsOutput, bool) bool) error {
	// If the supplied response is nil, then simulate an error
	if fake.LFOResponse == nil {
		return errors.New("ListFunctionsPages encountered an unexpected error: 1234")
	}

	// Loop through the slice of responses, invoking the supplied function
	for index, output := range fake.LFOResponse {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.LFOResponse)-1

		// Apply filtering to the supplied response
		// NOTE: I have not implemented this feature as our code does not require it.
		// To prevent unexpected cases, if the caller supplies an input other then
		// the "zero" input, the unit test fails.
		if input.FunctionVersion != nil || input.Marker != nil || input.MasterRegion != nil || input.MaxItems != nil {
			return errors.New("The unit test does not support a ListFunctionsInput other than 'zero' (no parameters)")
		}

		// Invoke our fn
		cont := fn(output, lastPage)

		// Shall we exit our loop?
		if !cont {
			break
		}
	}

	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Service Factory
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This structure simulates the AWS Service Factory by storing some pregenerated
// responses (that would come from AWS).
type fakeLambdaServiceFactory struct {
	RegionName string
	DRResponse *ec2.DescribeRegionsOutput
}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// This implementation of GetEC2InstanceService is limited to supporting DescribeRegions API
// only.
func (fsf fakeLambdaServiceFactory) GetEC2InstanceService(string) *EC2InstanceService {
	return &EC2InstanceService{
		Client: &fakeEC2Service{
			DRResponse: fsf.DRResponse,
		},
	}
}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) GetRDSInstanceService(string) *RDSInstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) GetS3Service() *S3Service {
	return nil
}

// Return a specialize LambdaService that returns a pre-canned response
func (fsf fakeLambdaServiceFactory) GetLambdaService(regionName string) *LambdaService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &LambdaService{
		Client: &fakeLambdaService{
			LFOResponse: lambdaFnsPerRegion[resolvedRegionName],
		},
	}
}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Don't need to implement
func (fsf fakeLambdaServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for LambdaFunctions
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestLambdaFunctions(t *testing.T) {
	// Describe all of our test cases: 1 failure and 4 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 4,
		}, {
			RegionName:    "us-east-2",
			ExpectedCount: 3,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 10,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 17,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeLambdaServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our Lambda Functions function
		actualCount := LambdaFunctions(sf, mon, c.AllRegions)

		// Did we expect an error?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else {
			if actualCount != c.ExpectedCount {
				t.Errorf("Error: LambdaFunctions returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
