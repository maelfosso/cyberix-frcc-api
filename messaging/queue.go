package messaging

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"cyberix.fr/frcc/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

type Queue struct {
	Client   *sqs.Client
	log      *zap.Logger
	mutex    sync.Mutex
	name     string
	url      *string
	waitTime time.Duration
}

type NewQueueOptions struct {
	Config   aws.Config
	Log      *zap.Logger
	Name     string
	WaitTime time.Duration
}

func NewQueue(opts NewQueueOptions) *Queue {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	return &Queue{
		Client:   sqs.NewFromConfig(opts.Config),
		log:      opts.Log,
		name:     opts.Name,
		waitTime: opts.WaitTime,
	}
}

func (q *Queue) getQueueURL(ctx context.Context) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.url != nil {
		return nil
	}

	output, err := q.Client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &q.name,
	})
	if err != nil {
		return err
	}
	q.url = output.QueueUrl

	return nil
}

func (q *Queue) Send(ctx context.Context, msg models.Message) error {
	if q.url == nil {
		if err := q.getQueueURL(ctx); err != nil {
			return err
		}
	}

	messageAsBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	messageAsString := string(messageAsBytes)

	_, err = q.Client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &messageAsString,
		QueueUrl:    q.url,
	})
	return err
}

func (q *Queue) Receive(ctx context.Context) (*models.Message, string, error) {
	if q.url == nil {
		if err := q.getQueueURL(ctx); err != nil {
			return nil, "", err
		}
	}

	output, err := q.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:        q.url,
		WaitTimeSeconds: int32(q.waitTime.Seconds()),
	})
	if err != nil {
		if strings.Contains(err.Error(), "context canceled") {
			return nil, "", nil
		}
		return nil, "", err
	}

	if len(output.Messages) == 0 {
		return nil, "", nil
	}

	var msg models.Message
	if err := json.Unmarshal([]byte(*output.Messages[0].Body), &msg); err != nil {
		return nil, "", err
	}

	return &msg, *output.Messages[0].ReceiptHandle, nil
}

func (q *Queue) Delete(ctx context.Context, receiptID string) error {
	if q.url == nil {
		if err := q.getQueueURL(ctx); err != nil {
			return err
		}
	}

	_, err := q.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      q.url,
		ReceiptHandle: &receiptID,
	})
	return err
}
