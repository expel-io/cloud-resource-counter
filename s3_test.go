package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/expel-io/cloud-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake S3 Buckets
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This simulates the minimal response from an AWS call
var fakeS3BucketsSlice = &s3.ListBucketsOutput{
	Buckets: []*s3.Bucket{
		{
			Name: aws.String("bucket1"),
		},
		{
			Name: aws.String("bucket2"),
		},
		{
			Name: aws.String("bucket3"),
		},
		{
			Name: aws.String("bucket4"),
		},
		{
			Name: aws.String("bucket5"),
		},
		{
			Name: aws.String("bucket6"),
		},
		{
			Name: aws.String("bucket7"),
		},
		{
			Name: aws.String("bucket8"),
		},
	},
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake S3 Service
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// To use this struct, the caller must supply a ListBucketsOutput struct. If
// it is missing, it will trigger the mock function to simulate an error from
// the corresponding function.
type fakeS3Service struct {
	s3iface.S3API
	LBResponse *s3.ListBucketsOutput
}

func (fs3 *fakeS3Service) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	// If there was no supplied response, then simulate a possible error
	if fs3.LBResponse == nil {
		return nil, errors.New("ListBuckets returns an unexpected error: 2345")
	}

	return fs3.LBResponse, nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Fake Service Factory
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// This structure simulates the AWS Service Factory by storing some pregenerated
// responses (that would come from AWS).
type fakeS3ServiceFactory struct {
	LBResponse *s3.ListBucketsOutput
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) Init() {}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetAccountIDService() *AccountIDService {
	return nil
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetEC2InstanceService(string) *EC2InstanceService {
	return nil
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	return nil
}

// Simply return our fake S3 Service
func (fsf fakeS3ServiceFactory) GetS3Service() *S3Service {
	return &S3Service{
		Client: &fakeS3Service{
			LBResponse: fsf.LBResponse,
		},
	}
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetLambdaService(string) *LambdaService {
	return nil
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetContainerService(string) *ContainerService {
	return nil
}

// Don't need to implement
func (fsf fakeS3ServiceFactory) GetLightsailService(string) *LightsailService {
	return nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for S3Buckets
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestS3Buckets(t *testing.T) {
	// Describe all of our test cases: 1 failure and 1 success
	cases := []struct {
		ExpectedCount int
		ExpectError   bool
	}{
		{
			ExpectedCount: 8,
		}, {
			ExpectError: true,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Construct a ListBucketsOutput object based on whether
		// we expect an error or not
		lbResponse := fakeS3BucketsSlice
		if c.ExpectError {
			lbResponse = nil
		}

		// Create our fake service factory
		sf := fakeS3ServiceFactory{
			LBResponse: lbResponse,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our S3 Buckets function
		actualCount := S3Buckets(sf, mon)

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
				t.Errorf("Error: S3Buckets returned %d; expected %d", actualCount, c.ExpectedCount)
			}
		}
	}
}
