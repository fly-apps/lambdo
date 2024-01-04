package broker

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/superfly/lambdo/internal/config"
	"github.com/superfly/lambdo/internal/fly"
	"github.com/superfly/lambdo/internal/logging"
	events "github.com/superfly/lambdo/internal/sqs"
	"go.uber.org/zap"
)

func SendToMachine(messages []types.Message) error {
	eventStrings := ""
	for k, m := range messages {
		if k == 0 {
			eventStrings += *m.Body
		} else {
			eventStrings += "," + *m.Body
		}
	}
	eventStringJson := fmt.Sprintf("[%s]", eventStrings)
	encodedJson := base64.StdEncoding.EncodeToString([]byte(eventStringJson))

	api := fly.NewApi(config.GetConfig().FlyToken)
	appName := config.GetConfig().FlyApp

	logging.GetLogger().Debug("creating Machine", zap.String("app-name", appName))

	// Each attempt iteration will try a new region
	regions := []string{"bos", "dfw", "den", "mia"}

	success := false

	for k, region := range regions {
		// TODO: Run the correct image (go vs js vs php base image)
		//       based on <something set by users>
		machine := fly.CreateMachineInput{
			AppName: config.GetConfig().FlyApp,
			Machine: fly.Machine{
				Region: region,
				Config: fly.MachineConfig{
					Image: "ubuntu:22.04",
					Env: map[string]string{
						"EVENTS_PATH": "/tmp/events.json",
					},
					Guest: fly.MachineSize{
						CpuCount: 2,
						RAM:      2048,
						Type:     "shared",
					},
					Processes: []fly.MachineProcess{
						{
							Cmd: []string{
								"cat",
								"/tmp/events.json",
							},
						},
						{
							Cmd: []string{
								"/bin/sleep",
								"3",
							},
						},
					},
					Files: []fly.MachineFile{
						{
							GuestPath: "/tmp/events.json",
							RawValue:  encodedJson,
						},
					},
					AutoDestroy: true,
				},
			},
		}

		m, err := api.CreateMachine(&machine)

		if err != nil {
			logging.GetLogger().Error("could not create Machine", zap.Error(err), zap.Int("attempt", k), zap.String("region", region))
			continue
		}

		success = true
		logging.GetLogger().Debug("created machine", zap.String("machine-id", m.Id))
		break
	}

	if success == false {
		return fmt.Errorf("could not create a Machine for this workload")
	} else {
		logging.GetLogger().Debug("machine created, deleting messages")

		// TODO: Delete machines if messages could not be deleted?
		for _, msg := range messages {
			delErr := events.DeleteMessage(msg)
			if delErr != nil {
				return fmt.Errorf("machines created, but could not delete messages: %w", delErr)
			}
		}
	}

	return nil
}
