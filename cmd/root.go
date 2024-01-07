package cmd

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/spf13/cobra"
	"github.com/superfly/lambdo/internal/broker"
	"github.com/superfly/lambdo/internal/logging"
	"github.com/superfly/lambdo/internal/sqs"
	"go.uber.org/zap"
	"os"
	"sync"
)

var rootCmd = &cobra.Command{
	Use:   "lambdo",
	Short: "Lambdo runs workloads based on events",
	Long: `This is like Lambda, but Flyier
Configuration should be set via environment variables. Possible values:
  required:
    LAMBDO_SQS_QUEUE_URL:         string, full sqs queue url
    AWS_*:                        Any needed AWS credential environment variables (region, key, secret, profile)
    LAMBDO_FLY_TOKEN              string, a valid Fly API token
    LAMBDO_FLY_REGION, FLY_REGION string, one of these must be set. FLY_REGION is already set when running in Fly
    LAMBDO_FLY_APP, FLY_APP_NAME  string, one of these must be set. FLY_APP_NAME is already set when running in Fly

  optional:
    LAMBDO_ENV:                   string, default: local
    LAMBDO_SQS_LONG_POLL_SECONDS: int,    default: 10
    LAMDBO_EVENTS_PER_MACHINE:    int,    default 5
`,
	Run: RunRootCommand,
}

var errors chan error

func Execute(ctx context.Context) error {
	errors = make(chan error)
	defer close(errors)

	go func() {
		for {
			select {
			case err := <-errors:
				logging.GetLogger().Error("broker error", zap.Error(err))
			case <-ctx.Done():
				logging.GetLogger().Info("Shutdown: No longer listening to error channel")
				return
			}
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func RunRootCommand(cmd *cobra.Command, args []string) {
	messages := make(chan []types.Message)
	defer close(messages)

	var brokerWorking sync.WaitGroup

	go func(ctx context.Context, m chan []types.Message) {
		for {
			select {
			case msgs := <-m:
				logging.GetLogger().Debug("messages received", zap.Any("messages", msgs))
				// todo: Should we send to broker in go routine to prevent
				//       holding up messages being processed (channel blocked until broker finishes)?
				brokerWorking.Add(1)
				err := broker.SendToMachine(msgs)
				if err != nil {
					errors <- err
				}
				brokerWorking.Done()
			case <-ctx.Done():
				logging.GetLogger().Info("Shutdown: no longer creating machines")
				return
			}
		}
	}(cmd.Context(), messages)

	// Listen for messages in SQS
	if err := sqs.Listen(cmd.Context(), messages); err != nil {
		logging.GetLogger().Error("SQS error", zap.Error(err))
		os.Exit(1)
	}

	logging.GetLogger().Info("Shutdown: waiting on broker to finish current job")

	brokerWorking.Wait()
}
