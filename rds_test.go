/******************************************************************************
Cloud Resource Counter
File: rds_test.go

Summary: The Unit Test for rds.
******************************************************************************/

package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/expel-io/aws-resource-counter/mock"
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
	// First page: 5 instances, 3 are available
	// Second page: 4 instances, 2 are available
	"us-east-2": []*rds.DescribeDBInstancesOutput{
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("backing-up"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("creating"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
			},
		},
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceStatus: aws.String("stopped"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("stopping"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
			},
		},
	},
	// AF-SOUTH-1 is an "opted in" region (Cape Town, Africa). We are going to
	// simply indicate that 3 instance exists here (only 1 running).
	"af-south-1": []*rds.DescribeDBInstancesOutput{
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceStatus: aws.String("upgrading"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("available"),
				},
				&rds.DBInstance{
					DBInstanceStatus: aws.String("backtracking"),
				},
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

	// Apply filtering to the supplied response
	// NOTE: I have not implemented this feature as our code does not require it.
	// To prevent unexpected cases, if the caller supplies an input other then
	// the "zero" input, the unit test fails.
	if input.DBInstanceIdentifier != nil || input.Filters != nil {
		return errors.New("The unit test does not support a DescribeDBInstancesInput other than 'zero' (no parameters)")
	}

	// Loop through the slice of responses, invoking the supplied function
	for index, output := range fake.DDBIResponse {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.DDBIResponse)-1

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

// Return our current region
func (fsf fakeRDSServiceFactory) GetCurrentRegion() string {
	return fsf.RegionName
}

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

// Don't need to implement
func (fsf fakeRDSServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Don't need to implement
func (fsf fakeRDSServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Don't need to implement
func (fsf fakeRDSServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for RDSInstances
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestRDSInstances(t *testing.T) {
	// Describe all of our test cases: 1 failure and 4 success cases
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

		// Invoke our RDS Counter function
		actualCount := RDSInstances(sf, mon, c.AllRegions)

		// Did we expect an error?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: RDSInstances returned %d; expected %d", actualCount, c.ExpectedCount)
		} else if mon.ProgramExited {
			t.Errorf("Unexpected Exit: The program unexpected exited with status code=%d", mon.ExitCode)
		}
	}
}
