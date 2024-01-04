package sqs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
)

var awsConfig aws.Config
var client *sqs.Client

func init() {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logging.GetLogger().Error("could not make aws config", zap.Error(err))
	} else {
		awsConfig = config
		client = sqs.NewFromConfig(awsConfig)
	}
}
