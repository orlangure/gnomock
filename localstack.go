// Package localstack provides a Gnomock Preset for localstack project
// (https://github.com/localstack/localstack). It allows to easily setup local
// AWS stack for testing
package localstack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/orlangure/gnomock"
)

// Localstack Preset exposes a number of ports, one for each AWS service
const (
	webPort = "web"

	APIGatewayPort       = "apigateway"
	CloudFormationPort   = "cloudformation"
	CloudWatchPort       = "cloudwatch"
	CloudWatchLogsPort   = "logs"
	CloudWatchEventsPort = "events"
	DynamoDBPort         = "dynamodb"
	DynamoDBStreamsPort  = "dynamodbstreams"
	EC2Port              = "ec2"
	ESPort               = "es"
	FirehosePort         = "firehose"
	IAMPort              = "iam"
	KinesisPort          = "kinesis"
	KMSPort              = "kms"
	LambdaPort           = "lambda"
	RedshiftPort         = "redshift"
	Route53Port          = "route53"
	S3Port               = "s3"
	SecretsManagerPort   = "secretsmanager"
	SESPort              = "ses"
	SNSPort              = "sns"
	SQSPort              = "sqs"
	SSMPort              = "ssm"
	STSPort              = "sts"
	StepFunctionsPort    = "stepfunctions"
)

// Preset creates a new localstack preset to use with gnomock.Start. See
// package docs for a list of exposed ports and services. It is legal to not
// provide any services using WithServices options, but in such case a new
// localstack container will be useless
func Preset(opts ...Option) gnomock.Preset {
	config := buildConfig(opts...)

	p := &localstack{
		services: config.services,
		s3Path:   config.s3Path,
	}

	return p
}

// Localstack is a Gnomock preset that exposes localstack functionality to spin
// up a number of AWS services locally
type localstack struct {
	services []Service

	s3Path string
}

// Image returns an image that should be pulled to create this container
func (p *localstack) Image() string {
	return "docker.io/localstack/localstack"
}

// Ports returns ports that should be used to access this container
func (p *localstack) Ports() gnomock.NamedPorts {
	return gnomock.NamedPorts{
		webPort: {Protocol: "tcp", Port: 8080},

		APIGatewayPort:       {Protocol: "tcp", Port: 4567},
		CloudFormationPort:   {Protocol: "tcp", Port: 4581},
		CloudWatchPort:       {Protocol: "tcp", Port: 4582},
		CloudWatchLogsPort:   {Protocol: "tcp", Port: 4569},
		CloudWatchEventsPort: {Protocol: "tcp", Port: 4587},
		DynamoDBPort:         {Protocol: "tcp", Port: 4569},
		DynamoDBStreamsPort:  {Protocol: "tcp", Port: 4570},
		EC2Port:              {Protocol: "tcp", Port: 4597},
		ESPort:               {Protocol: "tcp", Port: 4578},
		FirehosePort:         {Protocol: "tcp", Port: 4573},
		IAMPort:              {Protocol: "tcp", Port: 4593},
		KinesisPort:          {Protocol: "tcp", Port: 4568},
		KMSPort:              {Protocol: "tcp", Port: 4599},
		LambdaPort:           {Protocol: "tcp", Port: 4574},
		RedshiftPort:         {Protocol: "tcp", Port: 4577},
		Route53Port:          {Protocol: "tcp", Port: 4580},
		S3Port:               {Protocol: "tcp", Port: 4572},
		SecretsManagerPort:   {Protocol: "tcp", Port: 4584},
		SESPort:              {Protocol: "tcp", Port: 4579},
		SNSPort:              {Protocol: "tcp", Port: 4575},
		SQSPort:              {Protocol: "tcp", Port: 4576},
		SSMPort:              {Protocol: "tcp", Port: 4583},
		STSPort:              {Protocol: "tcp", Port: 4592},
		StepFunctionsPort:    {Protocol: "tcp", Port: 4585},
	}
}

// Options returns a list of options to configure this container
func (p *localstack) Options() []gnomock.Option {
	svcStrings := make([]string, len(p.services))
	for i, svc := range p.services {
		svcStrings[i] = string(svc)
	}

	svcEnv := strings.Join(svcStrings, ",")

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck(svcStrings)),
		gnomock.WithStartTimeout(time.Second * 60 * 2),
		gnomock.WithWaitTimeout(time.Second * 60),
		gnomock.WithEnv("SERVICES=" + svcEnv),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *localstack) healthcheck(services []string) gnomock.HealthcheckFunc {
	return func(c *gnomock.Container) (err error) {
		addr := fmt.Sprintf("http://%s/health", c.Address(webPort))

		res, err := http.Get(addr) //nolint:gosec
		if err != nil {
			return err
		}

		defer func() {
			closeErr := res.Body.Close()
			if err == nil && closeErr != nil {
				err = closeErr
			}
		}()

		var hr healthResponse

		decoder := json.NewDecoder(res.Body)

		err = decoder.Decode(&hr)
		if err != nil {
			return err
		}

		if len(hr.Services) < len(services) {
			return fmt.Errorf(
				"not enough active services: want %d got %d [%s]",
				len(services), len(hr.Services), hr.Services,
			)
		}

		for _, service := range services {
			status := hr.Services[service]
			if status != "running" {
				return fmt.Errorf("service '%s' is not running", service)
			}
		}

		return nil
	}
}

type healthResponse struct {
	Services map[string]string `json:"services"`
}

func (p *localstack) initf() gnomock.InitFunc {
	return func(c *gnomock.Container) error {
		for _, s := range p.services {
			if s == S3 {
				err := p.initS3(c)
				if err != nil {
					return fmt.Errorf("can't init s3 storage: %w", err)
				}
			}
		}

		return nil
	}
}
