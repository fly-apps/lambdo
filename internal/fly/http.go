package fly

import (
	"bytes"
	"fmt"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

// FlyRequest is an interface for API requests made to
// the Fly Machines API
type FlyRequest interface {
	ToRequest(token string) (*http.Request, error)
}

// StandardRequestHeaders adds standard request headers for Fly API
// requests, including the Authorization header with a Bearer token
func StandardRequestHeaders(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json")
}

// DoRequest runs an HTTP request, retrying any that time out
// due to possible "instability" in the Fly Machines API
func DoRequest(token string, r FlyRequest) (*http.Response, error) {
	defer func() {
		if r := recover(); r != nil {
			logging.GetLogger().Error("DoRequest panic", zap.Any("maybe-error", r))
		}
	}()

	req, err := r.ToRequest(token)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	result, err := doRequestWithRetries(req)
	if err != nil {
		return nil, err
	}

	// TODO: Result should never be nil here
	if result != nil && result.StatusCode > 299 {
		logging.GetLogger().Error(
			"API request to Fly returned unsuccessful status",
			zap.Int("status", result.StatusCode),
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
		)
	}

	return result, nil
}

func doRequestWithRetries(req *http.Request) (*http.Response, error) {
	defer func() {
		if r := recover(); r != nil {
			logging.GetLogger().Error("doRequestWithRetries panic", zap.Any("maybe-error", r))
		}
	}()

	var result *http.Response
	var err error
	for attempts := 1; attempts <= 5; attempts++ {
		logging.GetLogger().Debug(
			"making API request to Fly",
			zap.Int("attempt", attempts),
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
		)
		result, err = client.Do(req)
		if err != nil {
			if os.IsTimeout(err) {
				// Log body if available
				body := "nil"
				reqBody, reqBodyErr := req.GetBody()
				if reqBodyErr == nil {
					buf := new(bytes.Buffer)
					buf.ReadFrom(reqBody)
					body = buf.String()
				}

				logging.GetLogger().Debug("client timeout, retrying soon", zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.String("body", body))
				time.Sleep(time.Second * 1)
				continue
			}

			// If it's not a timeout, break out and return the error
			return nil, fmt.Errorf("http client error: %w", err)
		}

		// Deleting a machine may result in a 409 response
		// (even with the force flag)
		// TODO: Result should never be nil here
		if result != nil && result.StatusCode == http.StatusConflict {
			logging.GetLogger().Debug("conflict response, retrying soon", zap.String("method", req.Method), zap.String("url", req.URL.String()))
			time.Sleep(time.Second * 2)
			continue
		}

		// Deleting a machine may result in a 412 response
		// (even with the force flag)
		// TODO: Result should never be nil here
		if result != nil && result.StatusCode == http.StatusPreconditionFailed {
			logging.GetLogger().Debug("precondition failed response, retrying soon", zap.String("method", req.Method), zap.String("url", req.URL.String()))
			time.Sleep(time.Second * 2)
			continue
		}

		// 400 (bad request), preferably we know what exactly is wrong
		if result != nil && result.StatusCode == http.StatusBadRequest {
			// Log body if available
			body := "nil"
			buf := new(bytes.Buffer)
			_, bodyErr := buf.ReadFrom(result.Body)
			if bodyErr == nil {
				body = buf.String()
			}

			logging.GetLogger().Debug("bad request response", zap.String("method", req.Method), zap.String("url", req.URL.String()), zap.String("body", body))
		}

		logging.GetLogger().Debug(
			"made API request to Fly",
			zap.Int("attempt", attempts),
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Int("status", result.StatusCode),
		)
		return result, nil
	}

	return nil, err
}
