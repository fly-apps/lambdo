package sqs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/superfly/lambdo/internal/config"
)

func DeleteMessage(msg types.Message) error {
	_, err := client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(config.GetConfig().SQSQueueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		return fmt.Errorf("could not delete message: %w", err)
	}

	return nil
}
