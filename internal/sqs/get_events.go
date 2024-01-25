package sqs

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	cfg "github.com/superfly/lambdo/internal/config"
	"github.com/superfly/lambdo/internal/logging"
	"time"
)

type MessageHandler func([]types.Message)

func Listen(ctx context.Context, messages chan []types.Message) error {
	logging.GetLogger().Info("listening on SQS queue")

GETMSGS:
	for {
		select {
		case <-ctx.Done():
			logging.GetLogger().Info("Shutdown: No longer listening for messages")
			break GETMSGS
		default:
			logging.GetLogger().Debug("about to call sqs.ReceiveMessage")
			response, sqsErr := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(cfg.GetConfig().SQSQueueUrl),
				MaxNumberOfMessages:   int32(cfg.GetConfig().EventsPerMachine),   // max of 10
				WaitTimeSeconds:       int32(cfg.GetConfig().SQSLongPollSeconds), // long polling
				VisibilityTimeout:     30,                                        // POC queue defaults to 30, we mirror that here
				MessageAttributeNames: []string{"image", "size", "command"},
			})

			if sqsErr != nil {
				if !errors.Is(sqsErr, context.Canceled) {
					return fmt.Errorf("could not receive SQS messages: %w", sqsErr)
				} else {
					logging.GetLogger().Debug("ReceiveMessage stopped: context cancelled")
					return nil
				}

			}

			if len(response.Messages) > 0 {
				// Fire and forget from this function's point of view
				messages <- response.Messages
			}

			// Add time between calls if we don't long poll
			if cfg.GetConfig().SQSLongPollSeconds < 1 {
				time.Sleep(time.Millisecond * 250)
			}
		}
	}

	return nil
}
