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

type Event struct {
	Image string
	Body  string
}

type EventCollection struct {
	Events []*Event
}

func SendToMachine(messages []types.Message) error {
	api := fly.NewApi(config.GetConfig().FlyToken)
	appName := config.GetConfig().FlyApp
	// Each attempt iteration will try a new region
	regions := []string{"bos", "dfw", "den", "mia"}

	eventsPerMachine := map[string]*EventCollection{}

	// Group messages (events) by which image they require
	for _, m := range messages {
		image, err := findImage(m)
		if err != nil {
			logging.GetLogger().Warn("an event had no image", zap.String("error", err.Error()))
			continue
		}

		if _, ok := eventsPerMachine[image]; !ok {
			eventsPerMachine[image] = &EventCollection{}
		}

		eventsPerMachine[image].Events = append(eventsPerMachine[image].Events, &Event{
			Image: image,
			Body:  *m.Body,
		})
	}

	// Build JSON array of events, for each unique image
	// This assumes each event body is a
	// valid JSON string (lol)
	// This feels dumb af, but good enough for now
	for image, collection := range eventsPerMachine {
		eventStrings := ""
		for k, e := range collection.Events {
			if k == 0 {
				eventStrings += e.Body
			} else {
				eventStrings += "," + e.Body
			}
		}

		eventStringJson := fmt.Sprintf("[%s]", eventStrings)
		encodedJson := base64.StdEncoding.EncodeToString([]byte(eventStringJson))

		logging.GetLogger().Debug("creating Machine", zap.String("app-name", appName), zap.String("image", image))

		success := false

		for k, region := range regions {
			machine := fly.CreateMachineInput{
				AppName: config.GetConfig().FlyApp,
				Machine: fly.Machine{
					Region: region,
					Config: fly.MachineConfig{
						Image: image,
						Env: map[string]string{
							"EVENTS_PATH": "/tmp/events.json",
						},
						Guest: fly.MachineSize{
							CpuCount: 2,
							RAM:      2048,
							Type:     "shared",
						},
						//Processes: []fly.MachineProcess{
						//	{
						//		Cmd: []string{
						//			"cat",
						//			"/tmp/eventsPerMachine.json",
						//		},
						//	},
						//	{
						//		Cmd: []string{
						//			"/bin/sleep",
						//			"3",
						//		},
						//	},
						//},
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
				continue // try next region
			}

			success = true
			logging.GetLogger().Debug("created machine", zap.String("machine-id", m.Id))
			break // Break out of region retry loop
		}

		if success == false {
			logging.GetLogger().Error("could not create a Machine for this workload")
		} else {
			logging.GetLogger().Debug("machine created, deleting messages", zap.String("image", image))
		}
	}

	// TODO: Handle if messages could not be deleted (so it does not get re-tried?)
	// TODO: Delete messages as they are successfully added to a created machine? (all at once at the end
	//       could miss partial failure conditions)
	for _, msg := range messages {
		delErr := events.DeleteMessage(msg)
		if delErr != nil {
			return fmt.Errorf("machines created, but could not delete messages: %w", delErr)
		}
	}

	// We don't return an error when a machine is not created successfully
	return nil
}

func findImage(message types.Message) (string, error) {
	for k, v := range message.MessageAttributes {
		if k == "image" {
			return *v.StringValue, nil
		}
	}

	return "", fmt.Errorf("could not find an event image to run")
}
