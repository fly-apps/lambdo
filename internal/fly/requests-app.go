package fly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

/***********************************
 * Application Requests
***********************************/

/*************************
 * Create App Request
*************************/

type CreateAppRequest struct {
	Name    string `json:"app_name"`
	Org     string `json:"org_slug"`
	Network string `json:"network"`
}

func (r *CreateAppRequest) ToRequest(token string) (*http.Request, error) {
	j, err := json.Marshal(r)

	if err != nil {
		return nil, fmt.Errorf("could not encode App to JSON: %w", err)
	}

	url := fmt.Sprintf("%s/v1/apps", "https://api.machines.dev")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(j))

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Get App Request
*************************/

type GetAppRequest struct {
	App App
}

func (r *GetAppRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s", "https://api.machines.dev", r.App.Name)

	req, err := http.NewRequest(http.MethodGet, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}

/*************************
 * Delete App Request
*************************/

type DeleteAppRequest struct {
	App App
}

func (r *DeleteAppRequest) ToRequest(token string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/apps/%s", "https://api.machines.dev", r.App.Name)

	req, err := http.NewRequest(http.MethodDelete, uri, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create http request object: %w", err)
	}

	StandardRequestHeaders(req, token)

	return req, nil
}
