package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// DefaultRegion is used if the caller does not supply a region
// on the command line or the profile does not have a default
// region associated with it.
const DefaultRegion = "us-east-1"

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
// descriptions of all EC2 instances using an object that
// implements the Elastic Compute Cloud API interface.
type EC2InstanceService struct {
	Client ec2iface.EC2API
}

// InspectInstances takes an input filter specification (for the types of instances)
// and a function to evaluate a DescribeInstanceOutput struct. The supplied function
// can determine when to stop iterating through EC2 instances.
func (ec2i *EC2InstanceService) InspectInstances(input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return ec2i.Client.DescribeInstancesPages(input, fn)
}

// GetRegions returns the list of available regions for EC2 instances based on the
// set of input parameters.
func (ec2i *EC2InstanceService) GetRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	return ec2i.Client.DescribeRegions(input)
}

// ServiceFactory is the generic interface for our Cloud
// Service provider.
type ServiceFactory interface {
	Init()
	GetAccountIDService() *AccountIDService
	GetEC2InstanceService(string) *EC2InstanceService
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
	// TODO Use a map to store previously created Session associated with a given region?

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
