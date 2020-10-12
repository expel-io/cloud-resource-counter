/******************************************************************************
Cloud Resource Counter
File: lightsail_test.go

Summary: The Unit Test for lightsail.
******************************************************************************/

package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/lightsail/lightsailiface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Lightsail Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our list of accessible regions for the purpose of unit testing.
var lightsailRegions *lightsail.GetRegionsOutput = &lightsail.GetRegionsOutput{
	Regions: []*lightsail.Region{
		&lightsail.Region{
			Name: aws.String("us-east-1"),
		},
		&lightsail.Region{
			Name: aws.String("us-east-2"),
		},
		&lightsail.Region{
			Name: aws.String("eu-west-1"),
		},
	},
}

// This is our list of lightsail instances per region
var lightsailInstancesPerRegion = map[string]*lightsail.GetInstancesOutput{
	// US-EAST-1 simulates a region where there are two Lightsail instances:
	// one is Wordpress, the other is Node.js
	"us-east-1": &lightsail.GetInstancesOutput{
		Instances: []*lightsail.Instance{
			&lightsail.Instance{
				Name: aws.String("WordPress-1"),
			},
			&lightsail.Instance{
				Name: aws.String("Node-js-1"),
			},
		},
	},
	// US-EAST-2 has no instances...
	"us-east-2": &lightsail.GetInstancesOutput{},

	// EU-WEST-1 has 1 instance
	"eu-west-1": &lightsail.GetInstancesOutput{
		Instances: []*lightsail.Instance{
			&lightsail.Instance{
				Name: aws.String("WordPress-1"),
			},
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Lightsail Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our fake Lightsail Service that implements the AWS API for Lightsail
type fakeLightsailService struct {
	lightsailiface.LightsailAPI
	GRResponse  *lightsail.GetRegionsOutput
	GIOResponse *lightsail.GetInstancesOutput
}

// GetRegions fakes the standard Lightsail API of the same name.
func (fake *fakeLightsailService) GetRegions(input *lightsail.GetRegionsInput) (*lightsail.GetRegionsOutput, error) {
	// If the pre-canned regions response is nil, then simulate the API returning an error
	if fake.GRResponse == nil {
		return nil, errors.New("GetRegions encountered an unexpected error: 7531")
	}

	return fake.GRResponse, nil
}

// GetInstances fakes the standard Lightsail API of the same name.
func (fake *fakeLightsailService) GetInstances(input *lightsail.GetInstancesInput) (*lightsail.GetInstancesOutput, error) {
	// If the supplied response is nil, then simulate an error
	if fake.GIOResponse == nil {
		return nil, errors.New("GetInstance encountered an unexpected error: 02468")
	}

	return fake.GIOResponse, nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Lightsail Service Factory
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our fake Service Factory that implements a way to get a LightsailService.
type fakeLightsailServiceFactory struct {
	RegionName string
	GRResponse *lightsail.GetRegionsOutput
}

// Return our current region
func (fsf fakeLightsailServiceFactory) GetCurrentRegion() string {
	return fsf.RegionName
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetEC2InstanceService(string) *EC2InstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetS3Service() *S3Service {
	return nil
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Don't need to implement
func (fsf fakeLightsailServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Implement a way to return Lightsail regions and instances found in each
func (fsf fakeLightsailServiceFactory) GetLightsailService(regionName string) *LightsailService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &LightsailService{
		Client: &fakeLightsailService{
			GRResponse:  fsf.GRResponse,
			GIOResponse: lightsailInstancesPerRegion[resolvedRegionName],
		},
	}
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for LightsailInstances
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestLightsailInstances(t *testing.T) {
	// Describe all of our test cases: 2 failures and 4 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		GRResponse    *lightsail.GetRegionsOutput
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			GRResponse:    lightsailRegions,
			ExpectedCount: 2,
		}, {
			RegionName:    "us-east-2",
			GRResponse:    lightsailRegions,
			ExpectedCount: 0,
		}, {
			RegionName:    "eu-west-1",
			GRResponse:    lightsailRegions,
			ExpectedCount: 1,
		}, {
			RegionName:    "undefined-region",
			GRResponse:    lightsailRegions,
			ExpectedCount: 0,
		}, {
			AllRegions:    true,
			GRResponse:    lightsailRegions,
			ExpectedCount: 3,
		}, {
			AllRegions:  true,
			ExpectError: true,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeLightsailServiceFactory{
			RegionName: c.RegionName,
			GRResponse: c.GRResponse,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our LightsailInstances function
		actualCount := LightsailInstances(sf, mon, c.AllRegions)

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
				t.Errorf("Error: LightsailInstances returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
