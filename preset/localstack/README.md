# Gnomock Localstack ![Build](https://github.com/orlangure/gnomock-localstack/workflows/Build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock-localstack)](https://goreportcard.com/report/github.com/orlangure/gnomock-localstack)

Gnomock Localstack is a [Gnomock](https://github.com/orlangure/gnomock) preset
for running tests against AWS services locally, powered by
[Localstack](https://github.com/localstack/localstack) project. It allows
to setup a number of supported AWS services locally, run tests against
them, and tear them down easily.

See [Localstack](https://github.com/localstack/localstack) documentation for
more details.

### Testing against local S3

```go
package localstack_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/orlangure/gnomock"
	localstack "github.com/orlangure/gnomock-localstack"
)

func ExamplePreset_s3() {
	p := localstack.Preset(localstack.WithServices(localstack.S3))
	c, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(c) }()

	s3Endpoint := fmt.Sprintf("http://%s/", c.Address(localstack.S3Port))
	config := &aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(s3Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("a", "b", "c"),
	}
	sess, _ := session.NewSession(config)
	svc := s3.New(sess)

	_, _ = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("foo"),
	})

	out, _ := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("foo"),
	})
	fmt.Println("keys before:", *out.KeyCount)

	_, _ = svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader([]byte("this is a file")),
		Key:    aws.String("file"),
		Bucket: aws.String("foo"),
	})

	out, _ = svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("foo"),
	})
	fmt.Println("keys after:", *out.KeyCount)

	// Output:
	// keys before: 0
	// keys after: 1
}
```

### Testing against local SQS+SNS

```go
package localstack_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/orlangure/gnomock"
	localstack "github.com/orlangure/gnomock-localstack"
)

func ExamplePreset_sqs_sns() {
	p := localstack.Preset(
		localstack.WithServices(localstack.SNS, localstack.SQS),
	)
	c, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(c) }()

	endpoint := fmt.Sprintf("http://%s", c.Address(localstack.APIPort))

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials("a", "b", "c"),
	})

	sqsService := sqs.New(sess)
	snsService := sns.New(sess)

	_, _ = sqsService.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("my_queue"),
	})

	_, _ = snsService.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String("my_topic"),
	})

	queues, _ := sqsService.ListQueues(&sqs.ListQueuesInput{})
	fmt.Println("queues:", len(queues.QueueUrls))

	queueURL := queues.QueueUrls[0]

	topics, _ := snsService.ListTopics(&sns.ListTopicsInput{})
	fmt.Println("topics:", len(topics.Topics))

	topic := topics.Topics[0]

	_, _ = snsService.Subscribe(&sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		Endpoint: queueURL,
		TopicArn: topic.TopicArn,
	})

	_, _ = snsService.Publish(&sns.PublishInput{
		TopicArn: topic.TopicArn,
		Message:  aws.String("foobar"),
	})

	messages, _ := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: queueURL,
	})
	fmt.Println("messages:", len(messages.Messages))

	var msg map[string]string

	_ = json.Unmarshal([]byte(*messages.Messages[0].Body), &msg)
	fmt.Println("message:", msg["Message"])

	// Output:
	// queues: 1
	// topics: 1
	// messages: 1
	// message: foobar
}
```
