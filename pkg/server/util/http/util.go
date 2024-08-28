package httputil

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

func ProcessResponse(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	logger := logutil.GetLogger(ctx)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		logger.Info("Error sending request:", "error", err)
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("Error reading response body:", "error", err)
		return nil, nil, err
	}

	logger.Info("Response status:", "response status", resp.Status)
	logger.Info("Response body:", "body string", string(body))
	return resp, body, nil
}
