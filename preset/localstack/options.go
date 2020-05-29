package localstack

import (
	"encoding/json"
	"fmt"
)

// Option is an optional configuration of this Gnomock preset. Use available
// Options to configure the container
type Option func(*P)

// Service represents an AWS service that can be setup using localstack
type Service string

// UnmarshalJSON allows to unmarshal string into Service type
func (s *Service) UnmarshalJSON(bs []byte) error {
	var service string

	err := json.Unmarshal(bs, &service)
	if err != nil {
		return fmt.Errorf("invalid service '%s': %w", string(bs), err)
	}

	switch svc := Service(service); svc {
	case APIGateway,
		CloudFormation,
		CloudWatch,
		CloudWatchLogs,
		CloudWatchEvents,
		DynamoDB,
		DynamoDBStreams,
		EC2,
		ES,
		Firehose,
		IAM,
		Kinesis,
		KMS,
		Lambda,
		Redshift,
		Route53,
		S3,
		SecretsManager,
		SES,
		SNS,
		SQS,
		SSM,
		STS,
		StepFunctions:
		*s = svc
		return nil
	default:
		return fmt.Errorf("unknown service '%s'", svc)
	}
}

// These services are available in this Preset
const (
	APIGateway       Service = "apigateway"
	CloudFormation   Service = "cloudformation"
	CloudWatch       Service = "cloudwatch"
	CloudWatchLogs   Service = "logs"
	CloudWatchEvents Service = "events"
	DynamoDB         Service = "dynamodb"
	DynamoDBStreams  Service = "dynamodbstreams"
	EC2              Service = "ec2"
	ES               Service = "es"
	Firehose         Service = "firehose"
	IAM              Service = "iam"
	Kinesis          Service = "kinesis"
	KMS              Service = "kms"
	Lambda           Service = "lambda"
	Redshift         Service = "redshift"
	Route53          Service = "route53"
	S3               Service = "s3"
	SecretsManager   Service = "secretsmanager"
	SES              Service = "ses"
	SNS              Service = "sns"
	SQS              Service = "sqs"
	SSM              Service = "ssm"
	STS              Service = "sts"
	StepFunctions    Service = "stepfunctions"
)

// WithServices selects localstack services to spin up. It is OK to not select
// any services, but in such case the container will be useless
func WithServices(services ...Service) Option {
	return func(o *P) {
		o.Services = append(o.Services, services...)
	}
}

// WithVersion sets image version.
func WithVersion(version string) Option {
	return func(o *P) {
		o.Version = version
	}
}
