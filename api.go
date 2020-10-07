package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// DefaultRegion is used if the caller does not supply a region
// on the command line or the profile does not have a default
// region associated with it.
const DefaultRegion = "us-east-1"

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Abstract Services (hides details of Cloud Provider API)
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// AccountIDService is a struct that knows how get the AWS
// Account ID using an object that implements the Security
// Token Service API interface.
type AccountIDService struct {
	Client stsiface.STSAPI
}

// Account uses the supplied AccountIDService to invoke
// the associated GetCallerIdentity method on the struct's
// Client object.
func (aids *AccountIDService) Account() (string, error) {
	// Construct the input parameter
	input := &sts.GetCallerIdentityInput{}

	// Get the caller's identity
	result, err := aids.Client.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

// EC2InstanceService is a struct that knows how to get the
// descriptions of all EC2 instances as well as accessbile
// regions using an object that implements the Elastic
// Compute Cloud API interface.
type EC2InstanceService struct {
	Client ec2iface.EC2API
}

// InspectInstances takes an input filter specification (for the types of instances)
// and a function to evaluate a DescribeInstanceOutput struct. The supplied function
// can determine when to stop iterating through EC2 instances.
func (ec2i *EC2InstanceService) InspectInstances(input *ec2.DescribeInstancesInput,
	fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return ec2i.Client.DescribeInstancesPages(input, fn)
}

// GetRegions returns the list of available regions for EC2 instances based on the
// set of input parameters.
func (ec2i *EC2InstanceService) GetRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	return ec2i.Client.DescribeRegions(input)
}

// RDSInstanceService is a struct that knows how to get the
// descriptions of all RDS instances using an object that
// implements the Relational Database Service API interface.
type RDSInstanceService struct {
	Client rdsiface.RDSAPI
}

// InspectInstances takes an input filter specification (for the types of instances)
// and a function to evaluate a DescribeDBInstancesOutput struct. The supplied function
// can determine when to stop iterating through RDS instances.
func (rdsis *RDSInstanceService) InspectInstances(input *rds.DescribeDBInstancesInput,
	fn func(*rds.DescribeDBInstancesOutput, bool) bool) error {
	return rdsis.Client.DescribeDBInstancesPages(input, fn)
}

// S3Service is a struct that knows how to get all of the S3 buckets using an object
// that implements the Simple Storage Service API interface.
type S3Service struct {
	Client s3iface.S3API
}

// ListBuckets takes an input filter specification (for the types of S3 buckets) and
// returns a ListBucketsOutput struct.
func (s3s *S3Service) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return s3s.Client.ListBuckets(input)
}

// LambdaService is a struct that knows how to get all of the Lambda functions using
// an object that implements the Lambda API interface
type LambdaService struct {
	Client lambdaiface.LambdaAPI
}

// ListFunctions takes an input structure to identify specific lambda functions along
// with a function which is supplied a "page" of lambda functions.
func (ls *LambdaService) ListFunctions(input *lambda.ListFunctionsInput,
	fn func(*lambda.ListFunctionsOutput, bool) bool) error {
	return ls.Client.ListFunctionsPages(input, fn)
}

// ContainerService is a struct that knows how to get a list of all task definition
// and get a description of each one.
type ContainerService struct {
	Client ecsiface.ECSAPI
}

// ListTaskDefinitions takes an input specification (ListTaskDefinitionsInput) and
// a function that is invoked for each page of results (ListTaskDefinitionsOutput).
// This allows a caller to obtain a list of all task definitions.
func (cs *ContainerService) ListTaskDefinitions(input *ecs.ListTaskDefinitionsInput,
	fn func(output *ecs.ListTaskDefinitionsOutput, lastPage bool) bool) error {
	return cs.Client.ListTaskDefinitionsPages(input, fn)
}

// InspectTaskDefinition takes an input specification (DescribeTaskDefinitionInput)
// that describes a single task definition and returns information about it.
func (cs *ContainerService) InspectTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	return cs.Client.DescribeTaskDefinition(input)
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Abstract Service Factory (provides access to all Abstract Services)
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// ServiceFactory is the generic interface for our Cloud Service provider.
type ServiceFactory interface {
	Init()
	GetAccountIDService() *AccountIDService
	GetEC2InstanceService(string) *EC2InstanceService
	GetRDSInstanceService(string) *RDSInstanceService
	GetS3Service() *S3Service
	GetLambdaService(string) *LambdaService
	GetContainerService(string) *ContainerService
}

// AWSServiceFactory is a struct that holds a reference to
// an actual AWS Session object (pointer) and uses it to return
// other specialized services, such as the AccountIDService.
// It also accepts a profile name, overriding region and file
// to use to send trace information.
type AWSServiceFactory struct {
	Session     *session.Session
	ProfileName string
	RegionName  string
	TraceFile   *os.File
}

// Init initializes the AWS service factory by creating an
// initial AWS Session object (pointer). It inspects the profiles
// in the current user's directories and prepares the session for
// tracing (if requested).
func (awssf *AWSServiceFactory) Init() {
	// Create an initial configuration object (pointer)
	var config *aws.Config = aws.NewConfig()

	// Was a region specified by the user?
	if awssf.RegionName != "" {
		// Add it to the configuration
		config = config.WithRegion(awssf.RegionName)
	}

	// Was tracing specified by the user?
	if awssf.TraceFile != nil {
		// Enable logging of AWS Calls with Body
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)

		// Enable a logger function which writes to the Trace file
		config = config.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			fmt.Fprintln(awssf.TraceFile, args...)
		}))
	}

	// Construct our session Options object
	input := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           awssf.ProfileName,
		Config:            *config,
	}

	// Ensure that we have a session
	// It is not clear how a flawed session would manifest itself in this call
	// (as there is no error object returned). Perhaps a panic() is raised.
	sess := session.Must(session.NewSessionWithOptions(input))

	// Does this session have a region? If not, use the default region
	if *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(DefaultRegion)})
	}

	// Store the session in our struct
	awssf.Session = sess
}

// GetAccountIDService returns an instance of an AccountIDService associated
// with our session.
func (awssf *AWSServiceFactory) GetAccountIDService() *AccountIDService {
	return &AccountIDService{
		Client: sts.New(awssf.Session),
	}
}

// GetEC2InstanceService returns an instance of an EC2InstanceService associated
// with our session. The caller can supply an optional region name to contruct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetEC2InstanceService(regionName string) *EC2InstanceService {
	// Construct our service client
	var client ec2iface.EC2API
	if regionName == "" {
		client = ec2.New(awssf.Session)
	} else {
		client = ec2.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &EC2InstanceService{
		Client: client,
	}
}

// GetRDSInstanceService returns an instance of an RDSInstanceService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	// Construct our service client
	var client rdsiface.RDSAPI
	if regionName == "" {
		client = rds.New(awssf.Session)
	} else {
		client = rds.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &RDSInstanceService{
		Client: client,
	}
}

// GetS3Service returns an instance of an S3Service associated with the current session.
// There is currently no way to accept a different region name.
func (awssf *AWSServiceFactory) GetS3Service() *S3Service {
	return &S3Service{
		Client: s3.New(awssf.Session),
	}
}

// GetLambdaService returns an instance of a LambdaService associated with the our session.
// The caller can supply an optional region name to construct an instance associated with
// that region.
func (awssf *AWSServiceFactory) GetLambdaService(regionName string) *LambdaService {
	// Construct our service client
	var client lambdaiface.LambdaAPI
	if regionName == "" {
		client = lambda.New(awssf.Session)
	} else {
		client = lambda.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &LambdaService{
		Client: client,
	}
}

// GetContainerService returns an instance of a ContainerService associated with our session.
// The caller can supply an optional region name to construct an instance associated with
// that region.
func (awssf *AWSServiceFactory) GetContainerService(regionName string) *ContainerService {
	// Construct our service client
	var client ecsiface.ECSAPI
	if regionName == "" {
		client = ecs.New(awssf.Session)
	} else {
		client = ecs.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &ContainerService{
		Client: client,
	}
}
