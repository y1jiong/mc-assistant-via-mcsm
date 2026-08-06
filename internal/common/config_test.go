package common

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSendCommand(t *testing.T) {
	t.Parallel()

	config := Config{
		APIURL:     "https://mcsm.example.test/command",
		APIKey:     "secret",
		NodeID:     "node",
		InstanceID: "instance",
		httpClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			query := request.URL.Query()
			if got := query.Get("apikey"); got != "secret" {
				t.Errorf("apikey = %q, want secret", got)
			}
			if got := query.Get("daemonId"); got != "node" {
				t.Errorf("daemonId = %q, want node", got)
			}
			if got := query.Get("uuid"); got != "instance" {
				t.Errorf("uuid = %q, want instance", got)
			}
			if got := query.Get("command"); got != "say hello" {
				t.Errorf("command = %q, want say hello", got)
			}
			return response(http.StatusOK), nil
		})},
	}

	if err := config.SendCommand(context.Background(), "say hello"); err != nil {
		t.Fatalf("SendCommand() error = %v", err)
	}
}

func TestSendCommandReportsHTTPError(t *testing.T) {
	t.Parallel()

	config := Config{
		APIURL: "https://mcsm.example.test/command",
		httpClient: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return response(http.StatusServiceUnavailable), nil
		})},
	}
	if err := config.SendCommand(context.Background(), "list"); err == nil {
		t.Fatal("SendCommand() error = nil, want HTTP status error")
	}
}

func response(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
	}
}
