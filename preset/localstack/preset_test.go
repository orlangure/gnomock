package localstack_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset_s3(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"0.12.2", "0.13.1", "0.14.0", "2.3.0", "3.1.0"} {
		t.Run(version, testS3(version))
	}
}

func testS3(version string) func(*testing.T) {
	return func(t *testing.T) {
		p := localstack.Preset(
			localstack.WithServices(localstack.S3),
			localstack.WithVersion(version),
		)
		c, err := gnomock.Start(p, gnomock.WithTimeout(time.Minute*10))

		defer func() { require.NoError(t, gnomock.Stop(c)) }()

		require.NoError(t, err)

		s3Endpoint := fmt.Sprintf("http://%s/", c.Address(localstack.APIPort))
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "a",
					SecretAccessKey: "b",
					SessionToken:    "c",
				}, nil
			},
			)))

		require.NoError(t, err)

		svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &s3Endpoint
		})

		_, err = svc.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String("foo"),
		})
		require.NoError(t, err)

		out, err := svc.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String("foo"),
		})
		require.NoError(t, err)
		require.Empty(t, out.Contents)

		_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
			Body:   bytes.NewReader([]byte("this is a file")),
			Key:    aws.String("file"),
			Bucket: aws.String("foo"),
		})
		require.NoError(t, err)

		out, err = svc.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String("foo"),
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(out.Contents))
	}
}

func TestPreset_wrongS3Path(t *testing.T) {
	t.Parallel()

	p := localstack.Preset(
		localstack.WithServices(localstack.S3),
		localstack.WithS3Files("./invalid"),
	)
	c, err := gnomock.Start(p, gnomock.WithTimeout(time.Minute*10))
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read s3 initial files")
	require.NoError(t, gnomock.Stop(c))
}

func TestPreset_sqs_sns(t *testing.T) {
	t.Parallel()

	p := localstack.Preset(
		localstack.WithServices(localstack.SNS, localstack.SQS),
		localstack.WithVersion("3.1.0"),
	)
	c, err := gnomock.Start(p, gnomock.WithTimeout(time.Minute*10))

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	endpoint := fmt.Sprintf("http://%s", c.Address(localstack.APIPort))

	sqsService := sqs.New(sqs.Options{
		BaseEndpoint: &endpoint,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "a",
				SecretAccessKey: "b",
				SessionToken:    "c",
			}, nil
		}),
		Region: "us-east-1",
	})
	snsService := sns.New(sns.Options{
		BaseEndpoint: &endpoint,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "a",
				SecretAccessKey: "b",
				SessionToken:    "c",
			}, nil
		}),
		Region: "us-east-1",
	})

	_, err = sqsService.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String("my_queue"),
	})
	require.NoError(t, err)

	attrs, err := sqsService.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String("my_queue"),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	})
	require.NoError(t, err)
	require.Len(t, attrs.Attributes, 1)

	_, err = snsService.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String("my_topic"),
	})
	require.NoError(t, err)

	queues, err := sqsService.ListQueues(context.TODO(), &sqs.ListQueuesInput{})
	require.NoError(t, err)
	require.Len(t, queues.QueueUrls, 1)

	queueURL := queues.QueueUrls[0]
	queueARN := attrs.Attributes["QueueArn"]

	topics, err := snsService.ListTopics(context.TODO(), &sns.ListTopicsInput{})
	require.NoError(t, err)
	require.Equal(t, 1, len(topics.Topics))

	topic := topics.Topics[0]

	_, err = snsService.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		Endpoint: &queueARN,
		TopicArn: topic.TopicArn,
	})
	require.NoError(t, err)

	_, err = snsService.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: topic.TopicArn,
		Message:  aws.String("foobar"),
	})
	require.NoError(t, err)

	messages, err := sqsService.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:        &queueURL,
		WaitTimeSeconds: *aws.Int32(1),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(messages.Messages))

	var msg map[string]string

	err = json.Unmarshal([]byte(*messages.Messages[0].Body), &msg)
	require.NoError(t, err)
	require.Equal(t, "foobar", msg["Message"])
}
