package main

import (
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAwsServiceFactoryRegionResolution(t *testing.T) {
	// Build our test cases
	cases := []struct {
		SuppliedRegionName string
		ExpectedRegionName string
	}{
		{
			ExpectedRegionName: DefaultRegion,
		},
		{
			SuppliedRegionName: "us-west-2",
			ExpectedRegionName: "us-west-2",
		},
	}

	// Loop through our test cases...
	for _, c := range cases {
		// Create a new AWS Service Factory
		sf := &AWSServiceFactory{
			ProfileName: "non-existent-profile-name",
			RegionName:  c.SuppliedRegionName,
		}

		// Initialize it...
		sf.Init()

		// Let's inspect the generated session...
		sess := sf.Session

		// Does it match what we expected?
		if *sess.Config.Region != c.ExpectedRegionName {
			t.Errorf("Unexpected value for Region: expected %s, actual %s", c.ExpectedRegionName, *sess.Config.Region)
		}
	}
}

func TestAwsServiceFactoryTracing(t *testing.T) {
	// Create a Writer to stand in for our trace file
	builder := strings.Builder{}
	// Build our test cases
	cases := []struct {
		TraceWriter      io.Writer
		ExpectedLogLevel aws.LogLevelType
	}{
		{
			TraceWriter:      &builder,
			ExpectedLogLevel: aws.LogDebugWithHTTPBody,
		}, {},
	}

	// Loop through our test cases...
	for _, c := range cases {
		// Create a new AWS Service Factory
		sf := &AWSServiceFactory{
			TraceWriter: c.TraceWriter,
		}

		// Initialize it...
		sf.Init()

		// Let's inspect the generated session...
		sess := sf.Session

		// Does it have the correct log level
		if *sess.Config.LogLevel != c.ExpectedLogLevel {
			t.Errorf("Unexpected value for LogLevel: expected %v, actual %v", c.ExpectedLogLevel, *sess.Config.LogLevel)
		}
	}
}

func TestAwsServiceFactoryGetCurrentRegion(t *testing.T) {
	// Create a new session
	session, err := session.NewSession()
	if err != nil {
		t.Errorf("Unexpected error while creating a new session: %v", err)
	}

	// Create an AWS Service Factory
	sf := &AWSServiceFactory{
		Session: session,
	}

	// Get the current region
	region := sf.GetCurrentRegion()

	// Is it something other than empty string?
	if region != "" {
		t.Errorf("Unexpected current region: expected %s, actual %s", "", region)
	}
}

func TestAwsServiceFactoryGetAccountIDService(t *testing.T) {
	// Create a new session
	session, err := session.NewSession()
	if err != nil {
		t.Errorf("Unexpected error while creating a new session: %v", err)
	}

	// Create an AWS Service Factory
	sf := &AWSServiceFactory{
		Session: session,
	}

	// Get the desired service
	service := sf.GetAccountIDService()

	// Is the service nil?
	if service == nil {
		t.Errorf("No service returned for %s", "GetAccountIDService")
	}
}

func TestAwsServiceFactoryGetEC2InstanceService(t *testing.T) {
	// Create our test cases
	cases := []struct {
		RegionName string
	}{
		{},
		{
			RegionName: "us-west-1",
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a config for the region?
		var config = &aws.Config{}
		if c.RegionName != "" {
			config = config.WithRegion(c.RegionName)
		}

		// Create our test
		session, err := session.NewSession(config)
		if err != nil {
			t.Errorf("Unexpected error while creating a new session: %v", err)
		}

		// Create an AWS Service Factory
		sf := &AWSServiceFactory{
			Session: session,
		}

		// Get the desired service
		service := sf.GetEC2InstanceService(c.RegionName)

		// Is the service nil?
		if service == nil {
			t.Errorf("No service returned for %s", "GetEC2InstanceService")
		} else if service.Client != nil {
			// Convert to implementation type
			implType, ok := service.Client.(*ec2.EC2)
			if !ok {
				t.Errorf("Unexpected Client type: expected %v, actual %v", "*ec2.EC2", implType)
			} else if *implType.Config.Region != c.RegionName {
				t.Errorf("Unexpected value for Client.Config.Region: expected %s, actual %s", c.RegionName, *implType.Config.Region)
			}
		}
	}
}

func TestAwsServiceFactoryGetRDSInstanceService(t *testing.T) {
	// Create our test cases
	cases := []struct {
		RegionName string
	}{
		{},
		{
			RegionName: "us-west-1",
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a config for the region?
		var config = &aws.Config{}
		if c.RegionName != "" {
			config = config.WithRegion(c.RegionName)
		}

		// Create our test
		session, err := session.NewSession(config)
		if err != nil {
			t.Errorf("Unexpected error while creating a new session: %v", err)
		}

		// Create an AWS Service Factory
		sf := &AWSServiceFactory{
			Session: session,
		}

		// Get the desired service
		service := sf.GetRDSInstanceService(c.RegionName)

		// Is the service nil?
		if service == nil {
			t.Errorf("No service returned for %s", "GetRDSInstanceService")
		} else if service.Client != nil {
			// Convert to implementation type
			implType, ok := service.Client.(*rds.RDS)
			if !ok {
				t.Errorf("Unexpected Client type: expected %v, actual %v", "*rds.RDS", implType)
			} else if *implType.Config.Region != c.RegionName {
				t.Errorf("Unexpected value for Client.Config.Region: expected %s, actual %s", c.RegionName, *implType.Config.Region)
			}
		}
	}
}

func TestAwsServiceFactoryGetS3Service(t *testing.T) {
	// Create a new session
	session, err := session.NewSession()
	if err != nil {
		t.Errorf("Unexpected error while creating a new session: %v", err)
	}

	// Create an AWS Service Factory
	sf := &AWSServiceFactory{
		Session: session,
	}

	// Get the desired service
	service := sf.GetS3Service()

	// Is the service nil?
	if service == nil {
		t.Errorf("No service returned for %s", "GetS3Service")
	}
}

func TestAwsServiceFactoryGetLambdaService(t *testing.T) {
	// Create our test cases
	cases := []struct {
		RegionName string
	}{
		{},
		{
			RegionName: "us-west-1",
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a config for the region?
		var config = &aws.Config{}
		if c.RegionName != "" {
			config = config.WithRegion(c.RegionName)
		}

		// Create our test
		session, err := session.NewSession(config)
		if err != nil {
			t.Errorf("Unexpected error while creating a new session: %v", err)
		}

		// Create an AWS Service Factory
		sf := &AWSServiceFactory{
			Session: session,
		}

		// Get the desired service
		service := sf.GetLambdaService(c.RegionName)

		// Is the service nil?
		if service == nil {
			t.Errorf("No service returned for %s", "GetLambdaService")
		} else if service.Client != nil {
			// Convert to implementation type
			implType, ok := service.Client.(*lambda.Lambda)
			if !ok {
				t.Errorf("Unexpected Client type: expected %v, actual %v", "*lambda.Lambda", implType)
			} else if *implType.Config.Region != c.RegionName {
				t.Errorf("Unexpected value for Client.Config.Region: expected %s, actual %s", c.RegionName, *implType.Config.Region)
			}
		}
	}
}

func TestAwsServiceFactoryGetContainerService(t *testing.T) {
	// Create our test cases
	cases := []struct {
		RegionName string
	}{
		{},
		{
			RegionName: "us-west-1",
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a config for the region?
		var config = &aws.Config{}
		if c.RegionName != "" {
			config = config.WithRegion(c.RegionName)
		}

		// Create our test
		session, err := session.NewSession(config)
		if err != nil {
			t.Errorf("Unexpected error while creating a new session: %v", err)
		}

		// Create an AWS Service Factory
		sf := &AWSServiceFactory{
			Session: session,
		}

		// Get the desired service
		service := sf.GetContainerService(c.RegionName)

		// Is the service nil?
		if service == nil {
			t.Errorf("No service returned for %s", "GetContainerService")
		} else if service.Client != nil {
			// Convert to implementation type
			implType, ok := service.Client.(*ecs.ECS)
			if !ok {
				t.Errorf("Unexpected Client type: expected %v, actual %v", "*ecs.ECS", implType)
			} else if *implType.Config.Region != c.RegionName {
				t.Errorf("Unexpected value for Client.Config.Region: expected %s, actual %s", c.RegionName, *implType.Config.Region)
			}
		}
	}
}

func TestAwsServiceFactoryGetLightsailService(t *testing.T) {
	// Create our test cases
	cases := []struct {
		RegionName string
	}{
		{},
		{
			RegionName: "us-west-1",
		},
	}

	// Loop through the test cases
	for _, c := range cases {
		// Create a config for the region?
		var config = &aws.Config{}
		if c.RegionName != "" {
			config = config.WithRegion(c.RegionName)
		}

		// Create our test
		session, err := session.NewSession(config)
		if err != nil {
			t.Errorf("Unexpected error while creating a new session: %v", err)
		}

		// Create an AWS Service Factory
		sf := &AWSServiceFactory{
			Session: session,
		}

		// Get the desired service
		service := sf.GetLightsailService(c.RegionName)

		// Is the service nil?
		if service == nil {
			t.Errorf("No service returned for %s", "GetLightsailService")
		} else if service.Client != nil {
			// Convert to implementation type
			implType, ok := service.Client.(*lightsail.Lightsail)
			if !ok {
				t.Errorf("Unexpected Client type: expected %v, actual %v", "*lightsail.Lightsail", implType)
			} else if *implType.Config.Region != c.RegionName {
				t.Errorf("Unexpected value for Client.Config.Region: expected %s, actual %s", c.RegionName, *implType.Config.Region)
			}
		}
	}
}
