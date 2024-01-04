package fly

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type CreateAppInput struct {
	Name string
	Org  string
}

// CreateApp attempts to create an app, specifically
// using the Fly Machines API
func (api *Api) CreateApp(i *CreateAppInput) (*App, error) {
	app := i.Name
	org := i.Org

	if len(org) == 0 {
		org = OrgName()
	}

	req := &CreateAppRequest{
		Name:    app,
		Org:     org,
		Network: fmt.Sprintf("%s-net", i.Name),
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 299 {
		return nil, fmt.Errorf("could not create app, status: %d", response.StatusCode)
	}

	return &App{
		Name: app,
		Organization: Organization{
			Slug: org,
		},
	}, nil

	return nil, nil
}

type GetAppInput struct {
	Name string
}

// GetApp attempts to retrieve an application.
//
// An app that is not found returns an AppNotFoundError error
func (api *Api) GetApp(i *GetAppInput) (*App, error) {
	req := &GetAppRequest{
		App: App{
			Name: i.Name,
		},
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, AppNotFoundError{
			App: i.Name,
			Err: fmt.Errorf("fly returned 404 for app: %s", i.Name),
		}
	}

	// Other errors
	if response.StatusCode > 299 {
		return nil, fmt.Errorf("could not get app, status: %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("response body reading error: %w", err)
	}

	b := string(responseBody)
	if response.StatusCode > 399 {
		return nil, fmt.Errorf("invalid HTTP response '%d': %s", b)
	}

	logging.GetLogger().Debug("GetApp response", zap.String("body", b))

	a := &App{}
	err = json.Unmarshal(responseBody, a)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshall json: %w", err)
	}

	return a, nil
}

type DeleteAppInput struct {
	Name string
}

// DeleteApp deletes an application and
// any created Machines
func (api *Api) DeleteApp(i *DeleteAppInput) error {
	req := &DeleteAppRequest{
		App: App{
			Name: i.Name,
		},
	}

	response, err := DoRequest(api.Token, req)

	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	defer response.Body.Close()

	// Log "error" http responses, but otherwise ignore errors here
	if response.StatusCode > 299 {
		logging.GetLogger().Error("could not delete app", zap.Int("status", response.StatusCode))
	}

	return nil
}

// FindCreateApp will return an app if it already exists
// or attempts to create the app if the given app does
// not yet exist
func (api *Api) FindCreateApp(i *CreateAppInput) (*App, error) {
	var app *App
	var err error
	var appNotFoundError AppNotFoundError
	if app, err = api.GetApp(&GetAppInput{i.Name}); err != nil {
		if errors.As(err, &appNotFoundError) {
			app, err = api.CreateApp(i)
		} else {
			return nil, err
		}
	}

	// We have this err check here instead of nesting
	// another layer down after call to CreateApp()
	if err != nil {
		return nil, fmt.Errorf("could not create app: %s", i.Name)
	}

	return app, nil
}
