package broker

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
	Size  string
	Body  string
	Cmd   []string
	Meta  map[string]string
}

type EventCollection struct {
	Events []*Event
}

func SendToMachine(messages []types.Message) error {
	api := fly.NewApi(config.GetConfig().FlyToken)
	appName := config.GetConfig().FlyApp
	// Each attempt iteration will try a new region
	// TODO: Respect config.GetConfig().FlyRegion setting
	regions := []string{"bos", "dfw", "den", "mia"}

	eventsPerMachine := map[string]*EventCollection{}

	// Group messages (events) by which image, size, and command they require
	for _, m := range messages {
		image, imageErr := findAttribute("image", m)
		if imageErr != nil {
			logging.GetLogger().Warn("an event had no image", zap.String("error", imageErr.Error()))
			continue
		}

		size, sizeErr := findAttribute("size", m)

		if sizeErr != nil {
			logging.GetLogger().Warn("an event had no size, defaulting to performance-2x", zap.String("error", sizeErr.Error()))
			size = "performance-2x"
		}

		var cmd []string
		cmdString, cmdErr := findAttribute("command", m)

		if cmdErr != nil {
			logging.GetLogger().Warn("an event had no command, defaulting to no command", zap.String("error", cmdErr.Error()))
		}

		if jErr := json.Unmarshal([]byte(cmdString), &cmd); jErr != nil {
			logging.GetLogger().Warn("could not parse command, no machine will be created", zap.String("error", jErr.Error()))
			continue
		}

		// md5 of attributes that affect machine creation, so we can group like-events
		// into machines that run the same way
		eventsPerMachineKeyHash := md5.Sum([]byte(fmt.Sprintf("%s-%s-%s", image, size, cmdString)))
		eventsPerMachineKey := hex.EncodeToString(eventsPerMachineKeyHash[:])
		if _, ok := eventsPerMachine[eventsPerMachineKey]; !ok {
			eventsPerMachine[eventsPerMachineKey] = &EventCollection{}
		}

		eventsPerMachine[eventsPerMachineKey].Events = append(eventsPerMachine[eventsPerMachineKey].Events, &Event{
			Image: image,
			Body:  *m.Body,
			Size:  size,
			Cmd:   cmd,
			Meta:  map[string]string{"receipt": *m.ReceiptHandle},
		})
	}

	// Build JSON array of events, for each unique image
	// This assumes each event body is a
	// valid JSON string (lol)
	// This is dumb af, but good enough for now
	for _, collection := range eventsPerMachine {
		eventStrings := ""
		receipts := []string{}
		image := ""
		size := ""
		var cmd []string
		for k, e := range collection.Events {
			if k == 0 {
				eventStrings += e.Body
			} else {
				eventStrings += "," + e.Body
			}
			receipts = append(receipts, e.Meta["receipt"])

			// These get reset on every iteration, but in our scenario here,
			// they'll always get set to the same values since we segregated
			// them above. Just another code smell, no worries.
			image = e.Image
			size = e.Size
			cmd = e.Cmd
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
						/*
							Guest: fly.MachineSize{
								CpuCount: 2,
								RAM:      2048,
								Type:     "shared",
							},
						*/
						Size: size,
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

			if len(cmd) > 0 {
				machine.Machine.Config.Processes = []fly.MachineProcess{
					{
						Cmd: cmd,
					},
				}
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

			// TODO: Handle if messages could not be deleted (so it does not get re-tried?) - perhaps retry logic?
			for _, receipt := range receipts {
				delErr := events.DeleteMessage(receipt)
				if delErr != nil {
					logging.GetLogger().Error("machine crated but could not delete message", zap.Error(delErr))
				}
			}
		}
	}

	// We don't return an error when a machine fails to be created
	return nil
}

func findAttribute(attr string, message types.Message) (string, error) {
	for k, v := range message.MessageAttributes {
		if k == attr {
			return *v.StringValue, nil
		}
	}

	return "", fmt.Errorf("could not find an event %s", attr)
}
