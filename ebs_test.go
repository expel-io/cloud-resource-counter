/******************************************************************************
Cloud Resource Counter
File: ebs_test.go

Summary: The Unit Test for ebs.
******************************************************************************/

package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake EBS Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// This is our map of regions and the instances in each
var ebsVolumesPerRegion = map[string][]*ec2.DescribeVolumesOutput{
	// US-EAST-1 illustrates a case where DescribeVolumesPages returns 1 page
	// of results: 3 volumes, but only 2 are attached.
	"us-east-1": []*ec2.DescribeVolumesOutput{
		&ec2.DescribeVolumesOutput{
			Volumes: []*ec2.Volume{
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("some-instance-id"),
						},
					},
				},
				{
					Attachments: []*ec2.VolumeAttachment{},
				},
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("yet-another-instance-id"),
						},
					},
				},
			},
		},
	},

	// US-EAST-2 has 1 page of data: 7 Volumes in 3 reservations (1 spot
	// and 1 scheduled instance mixed in).
	"us-east-2": []*ec2.DescribeVolumesOutput{
		&ec2.DescribeVolumesOutput{
			Volumes: []*ec2.Volume{},
		},
	},

	// AF-SOUTH-1 is an "opted in" region (Cape Town, Africa). We are going to
	// simulate the case when DescribeVolumesPages returns three pages of
	// results. First page has 3 (all attached), second page has 3 (2 attached)
	// and the third page has 1 (attached).
	"af-south-1": []*ec2.DescribeVolumesOutput{
		&ec2.DescribeVolumesOutput{
			Volumes: []*ec2.Volume{
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("some-instance-id"),
						},
					},
				},
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("another-instance-id"),
						},
					},
				},
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("yet-another-instance-id"),
						},
					},
				},
			},
		},
		&ec2.DescribeVolumesOutput{
			Volumes: []*ec2.Volume{
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("and-another-instance-id"),
						},
					},
				},
				{
					Attachments: []*ec2.VolumeAttachment{},
				},
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("more-instance-id"),
						},
					},
				},
			},
		},
		&ec2.DescribeVolumesOutput{
			Volumes: []*ec2.Volume{
				{
					Attachments: []*ec2.VolumeAttachment{
						{
							InstanceId: aws.String("final-instance-id"),
						},
					},
				},
			},
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake EBS Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// To use this struct, the caller must supply a DescribeVolumesOutput slice
// and a DescribeRegionsOutput. If either is missing, it will trigger the mock
// functions to simulate an error from their corresponding functions.
type fakeEBSService struct {
	ec2iface.EC2API
	DVOResponse []*ec2.DescribeVolumesOutput
	DRResponse  *ec2.DescribeRegionsOutput
}

// Simulate the DescribeRegions function
func (fake *fakeEBSService) DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	// If the supplied response is nil, then simulate an error
	if fake.DRResponse == nil {
		return nil, errors.New("DescribeRegions encountered an unexpected error: 6789")
	}

	return fake.DRResponse, nil
}

// Simulate the DescribeVolumesPages function
func (fake *fakeEBSService) DescribeVolumesPages(input *ec2.DescribeVolumesInput,
	fn func(*ec2.DescribeVolumesOutput, bool) bool) error {
	// If the supplied response is nil, then simulate an error
	if fake.DVOResponse == nil {
		return errors.New("DescribeVolumes encountered an unexpected error: 1234")
	}

	// Apply filtering to the supplied response
	// NOTE: I have not implemented this feature as our code does not require it.
	// To prevent unexpected cases, if the caller supplies an input other then
	// the "zero" input, the unit test fails.
	if input.DryRun != nil || input.Filters != nil || input.MaxResults != nil || input.NextToken != nil || input.VolumeIds != nil {
		return errors.New("The unit test does not support a DescribeVolumesInput other than 'zero' (no parameters)")
	}

	// Loop through the slice, invoking the supplied function
	for index, output := range fake.DVOResponse {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.DVOResponse)-1

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
// responses (that would come from AWS)
type fakeEBSServiceFactory struct {
	RegionName string
	DRResponse *ec2.DescribeRegionsOutput
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) Init() {}

// Return our current region
func (fsf fakeEBSServiceFactory) GetCurrentRegion() string {
	return fsf.RegionName
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// Basic implementation
func (fsf fakeEBSServiceFactory) GetEC2InstanceService(regionName string) *EC2InstanceService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &EC2InstanceService{
		Client: &fakeEBSService{
			DVOResponse: ebsVolumesPerRegion[resolvedRegionName],
			DRResponse:  fsf.DRResponse,
		},
	}
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetRDSInstanceService(string) *RDSInstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetS3Service() *S3Service {
	return nil
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Don't need to implement
func (fsf fakeEBSServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for EBSVolumes
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestEBSVolumes(t *testing.T) {
	// Describe all of our test cases: 1 failure and 4 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 2,
		}, {
			RegionName:    "us-east-2",
			ExpectedCount: 0,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 6,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 8,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeEBSServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our EBSVolumes function
		actualCount := EBSVolumes(sf, mon, c.AllRegions)

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
				t.Errorf("Error: EBSVolumes returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
