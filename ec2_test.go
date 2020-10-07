package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake EC2 Region Data. This same data is also used for determining RDS
// Regions
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our list of accessible regions for the purpose of unit testing.
var ec2Regions *ec2.DescribeRegionsOutput = &ec2.DescribeRegionsOutput{
	Regions: []*ec2.Region{
		&ec2.Region{
			OptInStatus: aws.String("opt-in-not-required"),
			RegionName:  aws.String("us-east-1"),
		},
		&ec2.Region{
			OptInStatus: aws.String("opt-in-not-required"),
			RegionName:  aws.String("us-east-2"),
		},
		&ec2.Region{
			OptInStatus: aws.String("opted-in"),
			RegionName:  aws.String("af-south-1"),
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake EC2 Instance Data
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This is our map of regions and the instances in each
var ec2InstancesPerRegion = map[string][]*ec2.DescribeInstancesOutput{
	// US-EAST-1 illustrates a case where DescribeInstancesPages returns two pages of results.
	// First page: 2 different refervations (1 instance, then 2 instances [1 is spot])
	// Second page: 1 reservation (2 instances)
	"us-east-1": []*ec2.DescribeInstancesOutput{
		&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
					},
				},
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
						&ec2.Instance{
							InstanceLifecycle: aws.String("spot"),
						},
					},
				},
			},
		},
		&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
						&ec2.Instance{},
					},
				},
			},
		},
	},
	// US-EAST-2 has 1 page of data: 7 instances in 3 reservations (1 spot
	// and 1 scheduled instance mixed in).
	"us-east-2": []*ec2.DescribeInstancesOutput{
		&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
						&ec2.Instance{
							InstanceLifecycle: aws.String("scheduled"),
						},
						&ec2.Instance{},
					},
				},
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
						&ec2.Instance{},
						&ec2.Instance{},
					},
				},
				&ec2.Reservation{
					Instances: []*ec2.Instance{
						&ec2.Instance{},
						&ec2.Instance{
							InstanceLifecycle: aws.String("spot"),
						},
						&ec2.Instance{},
					},
				},
			},
		},
	},
	// AF-SOUTH-1 is an "opted in" region (Cape Town, Africa). We are going to
	// simply indicate that no instances exist here.
	"af-south-1": []*ec2.DescribeInstancesOutput{
		&ec2.DescribeInstancesOutput{},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake EC2 Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// To use this struct, the caller must supply a DescribeInstancesOutput slice
// and a DescribeRegionsOutput. If either is missing, it will trigger the mock
// functions to simulate an error from their corresponding functions.
type fakeEC2Service struct {
	ec2iface.EC2API
	DIPResponse []*ec2.DescribeInstancesOutput
	DRResponse  *ec2.DescribeRegionsOutput
}

// Simulate the DescribeRegions function
func (fake *fakeEC2Service) DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	// If the supplied response is nil, then simulate an error
	if fake.DRResponse == nil {
		return nil, errors.New("DescribeRegions encountered an unexpected error: 6789")
	}

	return fake.DRResponse, nil
}

// Helper function that converts a filter name ("instance-lifecycle") to a field name ("InstanceLifecycle")
func convertFilterNameToFieldName(filterName *string) string {
	// Split the string into parts separated by dashes
	parts := strings.Split(*filterName, "-")

	// Convert each part to an Uppercase version
	upperParts := Map(parts, strings.Title)

	return strings.Join(upperParts, "")
}

// Helper function that determines whether a reflected instance satisfied a single filter
func instanceSatisfiesFilter(reflectInstance *reflect.Value, filter *ec2.Filter) bool {
	// Convert our filter name to a field name
	fieldName := convertFilterNameToFieldName(filter.Name)

	// Is the name missing from the instance?
	var f reflect.Value
	if f = reflectInstance.FieldByName(fieldName); f.IsNil() {
		return false
	}

	// What is the value of the what the field name points to
	fieldNameValue := reflect.ValueOf(f.Interface())

	// Is the kind of value a Ptr? If not, get out now...
	if fieldNameValue.Kind() != reflect.Ptr {
		return false
	}

	// Dereference the pointer and get it as a string
	fieldValue := fieldNameValue.Elem().String()

	// Does this match one of the filter values?
	for _, value := range filter.Values {
		// Does it match?
		if *value == fieldValue {
			return true
		}
	}

	return false
}

// Helper function that determines whether an instance satifises the list of filters
func instanceSatisifiesFilters(instance *ec2.Instance, filters []*ec2.Filter) bool {
	// Is the input filters nil?
	if filters == nil {
		return true
	}

	// Perform reflection on the instance (struct)
	reflectedInstance := reflect.ValueOf(*instance)

	// Loop through the list of filters
	for _, filter := range filters {
		// Does the instance FAIL to satisfy the filter?
		if !instanceSatisfiesFilter(&reflectedInstance, filter) {
			return false
		}
	}

	return true
}

// Helper function that applies (limited) set of filtering criteria to the response
func applyDescribeInstancesInputFiltering(input *ec2.DescribeInstancesInput, output *ec2.DescribeInstancesOutput) *ec2.DescribeInstancesOutput {
	// Create a new DescribeInstancesOutput struct
	filteredOutput := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{},
	}

	// Loop through the reservations
	for _, origReservation := range output.Reservations {
		// Create a new Reservation to contain a list of filtered instances...
		filteredReservation := &ec2.Reservation{
			Instances: []*ec2.Instance{},
		}

		// Append it to the list of reservations...
		filteredOutput.Reservations = append(filteredOutput.Reservations, filteredReservation)

		// Loop through the instances...
		for _, instance := range origReservation.Instances {
			// Does the instance satisfy the filters?
			if instanceSatisifiesFilters(instance, input.Filters) {
				// Append the instance to the list
				filteredReservation.Instances = append(filteredReservation.Instances, instance)
			}
		}
	}

	return filteredOutput
}

// Simulate the DescribeInstancePages function
func (fake *fakeEC2Service) DescribeInstancesPages(input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	// If the supplied response is nil, then simulate an error
	if fake.DIPResponse == nil {
		return errors.New("DescribeInstancePages encountered an unexpected error: 1234")
	}

	// Loop through the slice, invoking the supplied function
	for index, output := range fake.DIPResponse {
		// Are we looking at the last "page" of our output?
		lastPage := index == len(fake.DIPResponse)-1

		// Apply filtering to the supplied response
		filteredOutput := applyDescribeInstancesInputFiltering(input, output)

		// Invoke our fn
		cont := fn(filteredOutput, lastPage)

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
type fakeEC2ServiceFactory struct {
	RegionName string
	DRResponse *ec2.DescribeRegionsOutput
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// Implement a way to return EC2 Regions and instances found in each
func (fsf fakeEC2ServiceFactory) GetEC2InstanceService(regionName string) *EC2InstanceService {
	// If the caller failed to specify a region, then use what is associated with our factory
	var resolvedRegionName string
	if regionName == "" {
		resolvedRegionName = fsf.RegionName
	} else {
		resolvedRegionName = regionName
	}

	return &EC2InstanceService{
		Client: &fakeEC2Service{
			DIPResponse: ec2InstancesPerRegion[resolvedRegionName],
			DRResponse:  fsf.DRResponse,
		},
	}
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetRDSInstanceService(string) *RDSInstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetS3Service() *S3Service {
	return nil
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Don't need to implement
func (fsf fakeEC2ServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for EC2Counts
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestEC2Counts(t *testing.T) {
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
			ExpectedCount: 7,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 0,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 11,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeEC2ServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our EC2 Counter function
		actualCount := EC2Counts(sf, mon, c.AllRegions)

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
				t.Errorf("Error: EC2Counts returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
