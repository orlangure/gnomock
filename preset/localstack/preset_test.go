package localstack_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/stretchr/testify/require"
)

func TestPreset_s3(t *testing.T) {
	t.Parallel()

	p := localstack.Preset(
		localstack.WithServices(localstack.S3),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	s3Endpoint := fmt.Sprintf("http://%s/", c.Address(localstack.APIPort))
	config := &aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(s3Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("a", "b", "c"),
	}
	sess, err := session.NewSession(config)
	require.NoError(t, err)

	svc := s3.New(sess)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("foo"),
	})
	require.NoError(t, err)

	out, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("foo"),
	})
	require.NoError(t, err)
	require.Empty(t, out.Contents)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader([]byte("this is a file")),
		Key:    aws.String("file"),
		Bucket: aws.String("foo"),
	})
	require.NoError(t, err)

	out, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("foo"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(out.Contents))
}

//nolint:funlen
func TestPreset_sqs_sns(t *testing.T) {
	t.Parallel()

	p := localstack.Preset(
		localstack.WithServices(localstack.SNS, localstack.SQS),
		localstack.WithVersion("0.11.0"),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	endpoint := fmt.Sprintf("http://%s", c.Address(localstack.APIPort))

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials("a", "b", "c"),
	})
	require.NoError(t, err)

	sqsService := sqs.New(sess)
	snsService := sns.New(sess)

	_, err = sqsService.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("my_queue"),
	})
	require.NoError(t, err)

	_, err = snsService.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String("my_topic"),
	})
	require.NoError(t, err)

	queues, err := sqsService.ListQueues(&sqs.ListQueuesInput{})
	require.NoError(t, err)
	require.Equal(t, 1, len(queues.QueueUrls))

	queueURL := queues.QueueUrls[0]

	topics, err := snsService.ListTopics(&sns.ListTopicsInput{})
	require.NoError(t, err)
	require.Equal(t, 1, len(topics.Topics))

	topic := topics.Topics[0]

	_, err = snsService.Subscribe(&sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		Endpoint: queueURL,
		TopicArn: topic.TopicArn,
	})
	require.NoError(t, err)

	_, err = snsService.Publish(&sns.PublishInput{
		TopicArn: topic.TopicArn,
		Message:  aws.String("foobar"),
	})
	require.NoError(t, err)

	messages, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: queueURL,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(messages.Messages))

	var msg map[string]string

	err = json.Unmarshal([]byte(*messages.Messages[0].Body), &msg)
	require.NoError(t, err)
	require.Equal(t, "foobar", msg["Message"])
}
