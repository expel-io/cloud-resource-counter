package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake RDS Instance Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our map of regions and the instances in each
var rdsInstancesPerRegion = map[string][]*rds.DescribeDBInstancesOutput{
	// US-EAST-1 illustrates a case where DescribeDBInstancesPages returns 1
	// page of NO results
	"us-east-1": []*rds.DescribeDBInstancesOutput{
		&rds.DescribeDBInstancesOutput{},
	},
	// US-EAST-2 illustrates a case where DescribeDBInstancesPages returns two pages of results.
	// First page: 3 instances
	// Second page: 2 instances
	"us-east-2": []*rds.DescribeDBInstancesOutput{
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{},
				&rds.DBInstance{},
				&rds.DBInstance{},
			},
		},
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{},
				&rds.DBInstance{},
			},
		},
	},
	// AF-SOUTH-1 is an "opted in" region (Cape Town, Africa). We are going to
	// simply indicate that 1 instance exists here.
	"af-south-1": []*rds.DescribeDBInstancesOutput{
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{},
			},
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake RDS Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// To use this struct, the caller must supply a DescribeDBInstances slice. If
// it is missing, it will trigger the mock function to simulate an error from
// the corresponding function.
type fakeRDSService struct {
	rdsiface.RDSAPI
	DDBIResponse []*rds.DescribeDBInstancesOutput
}

// Simulate the DescribeDBInstancesPages function
func (fake *fakeRDSService) DescribeDBInstancesPages(input *rds.DescribeDBInstancesInput, fn func(*rds.DescribeDBInstancesOutput, bool) bool) error {
	// If the supplied response is nil, then simulate an error
	if fake.DDBIResponse == nil {
		return errors.New("DescribeDBInstancesPages encountered an unexpected error: 1234")
	}

	// Loop through the slice of responses, invoking the supplied function
	for index, output := range fake.DDBIResponse {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.DDBIResponse)-1

		// Apply filtering to the supplied response
		// NOTE: I have not implemented this feature as our code does not require it.
		// To prevent unexpected cases, if the caller supplies an input other then
		// the "zero" input, the unit test.
		if input.DBInstanceIdentifier != nil || input.Filters != nil {
			return errors.New("The unit test does not support a DescribeDBInstancesInput other than 'zero' (no parameters)")
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
type fakeRDSServiceFactory struct {
	RegionName string
	DRResponse *ec2.DescribeRegionsOutput
}

// Don't need to implement
func (fsf fakeRDSServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeRDSServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// This implementation of GetEC2InstanceService is limited to supporting DescribeRegions API
// only.
func (fsf fakeRDSServiceFactory) GetEC2InstanceService(string) *EC2InstanceService {
	return &EC2InstanceService{
		Client: &fakeEC2Service{
			DRResponse: fsf.DRResponse,
		},
	}
}

// Implement a way to return a RDSInstanceService which is associated with the supplied
// region.
func (fsf fakeRDSServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &RDSInstanceService{
		Client: &fakeRDSService{
			DDBIResponse: rdsInstancesPerRegion[resolvedRegionName],
		},
	}
}

// Don't need to implement
func (fsf fakeRDSServiceFactory) GetS3Service() *S3Service {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for RDSInstances
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestRDSInstances(t *testing.T) {
	// Describe all of our test cases: 1 failure and 3 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 0,
		}, {
			RegionName:    "us-east-2",
			ExpectedCount: 5,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 1,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 6,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeRDSServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our RDS Counter method
		actualCount := RDSInstances(sf, mon, c.AllRegions)

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
				t.Errorf("Error: RDSInstances returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
