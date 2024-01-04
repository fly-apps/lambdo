package fly

import (
	"encoding/json"
	"fmt"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type CreateMachineInput struct {
	AppName string
	Machine Machine
}

func (api *Api) CreateMachine(i *CreateMachineInput) (*Machine, error) {
	req := &CreateMachineRequest{
		App: App{
			Name: i.AppName,
		},
		Machine: i.Machine,
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 299 {
		// Special handling for 422 error where Fly does not like our request
		if response.StatusCode == http.StatusUnprocessableEntity {
			loggedResponseBody := "no value yet"
			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				loggedResponseBody = fmt.Sprintf("error (response body reading): %s", err)
			} else {
				loggedResponseBody = string(responseBody)
			}
			return nil, fmt.Errorf("did not create machine, http status: %d, http body: %s", response.StatusCode, loggedResponseBody)
		} else {
			return nil, fmt.Errorf("did not create machine, http status: %d", response.StatusCode)
		}
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("response body reading error: %w", err)
	}

	m := &Machine{}
	err = json.Unmarshal(responseBody, m)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshall json: %w", err)
	}

	return m, nil
}

type GetMachineInput struct {
	AppName   string
	MachineId string
}

func (api *Api) GetMachine(i *GetMachineInput) (*Machine, error) {
	req := &GetMachineRequest{
		App: App{
			Name: i.AppName,
		},
		Machine: Machine{
			Id: i.MachineId,
		},
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer response.Body.Close()

	// Incorrect Machine ID's might return a 400 instead of a 404
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusBadRequest {
		return nil, MachineNotFoundError{
			App:     i.AppName,
			Machine: i.MachineId,
			Err:     err,
		}
	}

	// Other errors
	if response.StatusCode > 299 {
		return nil, fmt.Errorf("could not get machine, http status: %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)

	b := string(responseBody)
	if response.StatusCode > 399 {
		return nil, fmt.Errorf("invalid HTTP response '%d': %s", b)
	}

	if err != nil {
		return nil, fmt.Errorf("response body reading error: %w", err)
	}

	logging.GetLogger().Debug("GetMachine response", zap.String("body", b))

	m := &Machine{}
	err = json.Unmarshal(responseBody, m)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshall json: %w", err)
	}

	return m, nil
}

type ListMachinesInput struct {
	AppName string
}

func (api *Api) ListMachines(i *ListMachinesInput) (*ListMachinesResponse, error) {
	req := &ListMachinesRequest{
		App: App{
			Name: i.AppName,
		},
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 299 {
		return nil, fmt.Errorf("could not list machines, http status: %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("response body reading error: %w", err)
	}

	m := &ListMachinesResponse{}
	err = json.Unmarshal(responseBody, &m.Machines)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshall json: %w", err)
	}

	return m, nil
}

/* TODO: Implement this. It's similar (equivalent to?)
         the CreateMachine API call
type UpdateMachineInput struct {

}

func (api *Api) UpdateMachine() error {

}
*/

type DeleteMachineInput struct {
	AppName   string
	MachineId string
	Force     bool
}

func (api *Api) DeleteMachine(i *DeleteMachineInput) error {
	req := &DeleteMachineRequest{
		App: App{
			Name: i.AppName,
		},
		Machine: Machine{
			Id: i.MachineId,
		},
		Force: i.Force,
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	if response != nil {
		// Log "error" http responses, but otherwise ignore errors here
		if response.StatusCode > 299 {
			logging.GetLogger().Error("could not delete machine", zap.Int("status", response.StatusCode))
		}
	} else {
		logging.GetLogger().Error("Delete Machine: We should have a response object here, but we do not")
	}

	return nil
}

type StartMachineInput struct {
	AppName   string
	MachineId string
}

func (api *Api) StartMachine(i *StartMachineInput) error {
	req := &StartMachineRequest{
		App: App{
			Name: i.AppName,
		},
		Machine: Machine{
			Id: i.MachineId,
		},
	}

	_, err := DoRequest(api.Token, req)

	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	return nil
}

type StopMachineInput struct {
	AppName   string
	MachineId string
}

func (api *Api) StopMachine(i *StopMachineInput) error {
	req := &StopMachineRequest{
		App: App{
			Name: i.AppName,
		},
		Machine: Machine{
			Id: i.MachineId,
		},
	}

	_, err := DoRequest(api.Token, req)

	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	return nil
}

// WaitForMachine waits for a newly created Machine
// to become available
func (api *Api) WaitForMachine(i *GetMachineInput) error {
	// Total wait time ~5 minutes (should only need a minute or 2)
	ticker := time.NewTicker(2 * time.Second)
	totalAttempts := 0
	attemptsAllowed := 150

	for {
		select {
		case <-ticker.C:
			e, err := api.GetMachine(i)
			if err != nil {
				ticker.Stop()
				return fmt.Errorf("could not get machine: %w", err)
			}

			if e.IsInitialized() {
				ticker.Stop()
				return nil
			}

			totalAttempts++
			if totalAttempts >= attemptsAllowed {
				ticker.Stop()
				return fmt.Errorf("too many GetMachine attempts")
			}
		}
	}
}
