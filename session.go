package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// EstablishAwsSession does stuff...
func EstablishAwsSession() *session.Session {
	input := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profileName,
	}

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

	return sess
}
