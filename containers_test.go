package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Task Definitions Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// For our tests, we are combining the list of task definition ARNs with the
// descriptions of each.
type TaskInfo struct {
	ListOutputs       []*ecs.ListTaskDefinitionsOutput
	DescribeOutputMap map[string]*ecs.DescribeTaskDefinitionOutput
}

// This is our map of regions and the task definitions in each
var taskDefinitionsPerRegion = map[string]*TaskInfo{
	// US-EAST-1 simulations the case when more than one page of ListTaskDefinitionsOutput is
	// returned. In this case, there are 2 task definitions in the first page and 1 task
	// definition in the other page. Each of the 3 task definitions are listed as well. The
	// first task definition uses two containers, each with different images. The second task
	// definition uses one container, with a single image. The third task has a single container
	// but uses the same image as the first task. In total, there are 3 task definitions, 4
	// containers, but with only 3 unique container images.
	"us-east-1": &TaskInfo{
		ListOutputs: []*ecs.ListTaskDefinitionsOutput{
			&ecs.ListTaskDefinitionsOutput{
				TaskDefinitionArns: []*string{
					aws.String("some-long-name:task-definition/family:1"),
					aws.String("some-long-name:task-definition/family:2"),
				},
			},
			&ecs.ListTaskDefinitionsOutput{
				TaskDefinitionArns: []*string{
					aws.String("some-long-name:task-definition/otherfamily:1"),
				},
			},
		},
		DescribeOutputMap: map[string]*ecs.DescribeTaskDefinitionOutput{
			"some-long-name:task-definition/family:1": &ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						&ecs.ContainerDefinition{
							Image: aws.String("image1"),
						},
						&ecs.ContainerDefinition{
							Image: aws.String("image2"),
						},
					},
				},
			},
			"some-long-name:task-definition/family:2": &ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						&ecs.ContainerDefinition{
							Image: aws.String("image3"),
						},
					},
				},
			},
			"some-long-name:task-definition/otherfamily:1": &ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						&ecs.ContainerDefinition{
							Image: aws.String("image1"),
						},
					},
				},
			},
		},
	},
	// US-EAST-2 simulates the case when a single page of ListTaskDefinitionOutput is returned
	// to the caller. There are two task definitions. Each has a single container image, which
	// are different. In total, there are 2 task definitions, 2 cotnainers and 2 unique container
	// images.
	"us-east-2": &TaskInfo{
		ListOutputs: []*ecs.ListTaskDefinitionsOutput{
			&ecs.ListTaskDefinitionsOutput{
				TaskDefinitionArns: []*string{
					aws.String("some-long-name:task-definition/family:1"),
					aws.String("some-long-name:task-definition/anotherfamily:1"),
				},
			},
		},
		DescribeOutputMap: map[string]*ecs.DescribeTaskDefinitionOutput{
			"some-long-name:task-definition/family:1": &ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						&ecs.ContainerDefinition{
							Image: aws.String("image2"),
						},
					},
				},
			},
			"some-long-name:task-definition/anotherfamily:1": &ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						&ecs.ContainerDefinition{
							Image: aws.String("image4"),
						},
					},
				},
			},
		},
	},
	// AF-SOUTH-1 indicates that no tasks were defined for this regino.
	"af-south-1": &TaskInfo{
		ListOutputs: []*ecs.ListTaskDefinitionsOutput{
			&ecs.ListTaskDefinitionsOutput{},
		},
	},
	// AF-SOUTH-2 simulates a flawed case--what would happen if DescribeTaskDefinition
	// ever returned a failure? We simulate that by indicating that there is a Task Definition
	// ARN for which there is no description.
	"af-south-2": &TaskInfo{
		ListOutputs: []*ecs.ListTaskDefinitionsOutput{
			&ecs.ListTaskDefinitionsOutput{
				TaskDefinitionArns: []*string{
					aws.String("this-is-a-non-existent-task-arn-which-triggers-failure"),
				},
			},
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Container Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This structure simulates a ContainerService by providing both an array of
// pre-canned outputs (for the ListTaskDefinitions method) as well as a map of
// pre-canned outputs (for the InspectTaskDefinition method).
type fakeContainerService struct {
	ecsiface.ECSAPI
	TaskInfo *TaskInfo
}

// Implement the ListTaskDefinitionsPages method by returning the pre-canned array
// of ListTaskDefinitionsOutput structs.
func (fake *fakeContainerService) ListTaskDefinitionsPages(input *ecs.ListTaskDefinitionsInput,
	fn func(page *ecs.ListTaskDefinitionsOutput, lastPage bool) bool) error {
	// If there is no TaskInfo, simulate an error...
	if fake.TaskInfo == nil {
		return errors.New("ListTaskDefinitionsPages encountered an unexpected error: 2468")
	}

	// Loop through the slice of responses, invoking the supplied function
	for index, output := range fake.TaskInfo.ListOutputs {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.TaskInfo.ListOutputs)-1

		// Apply filtering to the supplied response
		// NOTE: I have not implemented this feature as our code does not require it.
		// To prevent unexpected cases, if the caller supplies an input other then
		// the "zero" input, the unit test fails.
		if input.FamilyPrefix != nil || input.MaxResults != nil || input.NextToken != nil || input.Sort != nil || input.Status != nil {
			return errors.New("The unit test does not support a ListTaskDefinitionsInput other than 'zero' (no parameters)")
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

// Implement the DescribeTaskDefinition method by returning a pre-canned response
// keyed by the task definition ARN.
func (fake *fakeContainerService) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	// We ensure that the input does not contain unexpected fields
	if input.Include != nil {
		return nil, errors.New("The unit test does not support a DescribeTaskDefinitionInput that contains anything other than a TaskDefinition")
	}

	// Get the response keyed by the task ARN
	response, ok := fake.TaskInfo.DescribeOutputMap[*input.TaskDefinition]

	// If no match, return an error
	if !ok {
		return nil, errors.New("DescribeTaskDefinition encountered an unexpected error: 9876")
	}

	return response, nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Service Factory
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This structure simulates the AWS Service Factory by storing some pregenerated
// responses (that would come from AWS).
type fakeCntrServiceFactory struct {
	RegionName string
	DRResponse *ec2.DescribeRegionsOutput
}

// Don't need to implement
func (fsf fakeCntrServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeCntrServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// This implementation of GetEC2InstanceService is limited to supporting DescribeRegions API
// only.
func (fsf fakeCntrServiceFactory) GetEC2InstanceService(string) *EC2InstanceService {
	return &EC2InstanceService{
		Client: &fakeEC2Service{
			DRResponse: fsf.DRResponse,
		},
	}
}

// Don't need to implement
func (fsf fakeCntrServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeCntrServiceFactory) GetS3Service() *S3Service {
	return nil
}

// Don't need to implement
func (fsf fakeCntrServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Implement a way to return a ContainerService instance associated with a specific
// region
func (fsf fakeCntrServiceFactory) GetContainerService(regionName string) *ContainerService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &ContainerService{
		Client: &fakeContainerService{
			TaskInfo: taskDefinitionsPerRegion[resolvedRegionName],
		},
	}
}

// Don't need to implement
func (fsf fakeCntrServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for UniqueContainerImages
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestUniqueContainerImages(t *testing.T) {
	// Describe all of our test cases: 2 failures and 4 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 3,
		}, {
			RegionName:    "us-east-2",
			ExpectedCount: 2,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 0,
		}, {
			RegionName:  "af-south-2",
			ExpectError: true,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 4,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeCntrServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our UniqueContainerImages function
		actualCount := UniqueContainerImages(sf, mon, c.AllRegions)

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
				t.Errorf("Error: UniqueContainerImages returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
