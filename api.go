package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// CallerIdentityService blah, blah, blah...
type CallerIdentityService struct {
	Client stsiface.STSAPI
}

// Account does shit...
func (cis *CallerIdentityService) Account() (string, error) {
	// Construct the input parameter
	input := &sts.GetCallerIdentityInput{}

	// Get the caller's identity
	result, err := cis.Client.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

// AWSServiceFactory is awesome
type AWSServiceFactory struct {
	Session *session.Session
}

// Init does stuff...
func (awssf *AWSServiceFactory) Init() {
	// Construct our session Options object
	input := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profileName,
	}

	// Create an initial configuration
	var config *aws.Config = aws.NewConfig()

	// Was region specified?
	if regionName != "" {
		config = config.WithRegion(regionName)
	}

	// Was tracing specified?
	if traceFile != nil {
		// Enable logging of AWS Calls with Body
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)

		// Enable a logger function which writes to the Trace file
		config = config.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			fmt.Fprintln(traceFile, args...)
		}))
	}

	// Attach the config
	input.Config = *config

	// Ensure that we have a session
	sess := session.Must(session.NewSessionWithOptions(input))

	// Does this session have a region? If not, we should specify US-EAST-1 as a default
	if *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}

	// Store the session in our struct
	awssf.Session = sess
}

// GetCallerIdentityService ...
func (awssf *AWSServiceFactory) GetCallerIdentityService() *CallerIdentityService {
	return &CallerIdentityService{
		Client: sts.New(awssf.Session),
	}
}
