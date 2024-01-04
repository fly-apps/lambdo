package fly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
	"net/http"
)

/***********************************
 * Machine Requests
***********************************/

/*************************
 * Create Machine Request
*************************/

type CreateMachineRequest struct {
	App     App
	Machine Machine
}

func (r *CreateMachineRequest) ToRequest(token string) (*http.Request, error) {
	j, err := json.Marshal(r.Machine)

	if err != nil {
		return nil, fmt.Errorf("could not encode Machine to JSON: %w", err)
	}

	logging.GetLogger().Debug("create machine request", zap.ByteString("body", j))

	uri := fmt.Sprintf("%s/v1/apps/%s/machines", "https://api.machines.dev", r.App.Name)
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(j))

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Get Machine Request
*************************/

type GetMachineRequest struct {
	App     App
	Machine Machine
}

func (r *GetMachineRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s/machines/%s", "https://api.machines.dev", r.App.Name, r.Machine.Id)

	req, err := http.NewRequest(http.MethodGet, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * List Machines Request
*************************/

type ListMachinesRequest struct {
	App App
}

type ListMachinesResponse struct {
	Machines []Machine
}

func (r *ListMachinesRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s/machines", "https://api.machines.dev", r.App.Name)
	req, err := http.NewRequest(http.MethodGet, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Update Machine Request
*************************/

type UpdateMachineRequest struct {
	App     App
	Machine Machine
}

func (r *UpdateMachineRequest) ToRequest(token string) (*http.Request, error) {
	j, err := json.Marshal(r.Machine)

	if err != nil {
		return nil, fmt.Errorf("could not encode Machine to JSON: %w", err)
	}

	uri := fmt.Sprintf("%s/v1/apps/%s/machines/%s", "https://api.machines.dev", r.App.Name, r.Machine.Id)
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(j))

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Delete Machine Request
*************************/

type DeleteMachineRequest struct {
	App     App
	Machine Machine
	Force   bool
}

func (r *DeleteMachineRequest) ToRequest(token string) (*http.Request, error) {
	var force string
	if r.Force {
		force = "?kill=true"
	} else {
		force = ""
	}

	uri := fmt.Sprintf("%s/v1/apps/%s/machines/%s%s", "https://api.machines.dev", r.App.Name, r.Machine.Id, force)

	req, err := http.NewRequest(http.MethodDelete, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Start Machine Request
*************************/

type StartMachineRequest struct {
	App     App
	Machine Machine
}

func (r *StartMachineRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s/machines/%s/start", "https://api.machines.dev", r.App.Name, r.Machine.Id)
	req, err := http.NewRequest(http.MethodPost, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Stop Machine Request
*************************/

type StopMachineRequest struct {
	App     App
	Machine Machine
}

func (r *StopMachineRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s/machines/%s/stop", "https://api.machines.dev", r.App.Name, r.Machine.Id)
	req, err := http.NewRequest(http.MethodPost, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}
