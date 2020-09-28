package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// DefaultRegion is used if the caller does not supply a region
// on the command line or the profile does not have a default
// region associated with is.
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
func (cis *AccountIDService) Account() (string, error) {
	// Construct the input parameter
	input := &sts.GetCallerIdentityInput{}

	// Get the caller's identity
	result, err := cis.Client.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

// AWSServiceFactory is a struct that holds a reference to
// an actual AWS Session object (pointer) and uses it to return
// other specialized services, such as the AccountIDService.
type AWSServiceFactory struct {
	Session *session.Session
}

// Init initializes the AWS service factory by creating an
// initial AWS Session object (pointer). It inspects the profiles
// in the current user's directories and prepares the session for
// tracing (if requested).
func (awssf *AWSServiceFactory) Init() {
	// Create an initial configuration object (pointer)
	var config *aws.Config = aws.NewConfig()

	// Was a region specified by the user?
	if regionName != "" {
		// Add it to the configuration
		config = config.WithRegion(regionName)
	}

	// Was tracing specified by the user?
	if traceFile != nil {
		// Enable logging of AWS Calls with Body
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)

		// Enable a logger function which writes to the Trace file
		config = config.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			fmt.Fprintln(traceFile, args...)
		}))
	}

	// Construct our session Options object
	input := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profileName,
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
