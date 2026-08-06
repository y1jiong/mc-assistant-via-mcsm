package common

import (
	"context"
	"fmt"
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
			if got := request.Header.Get("Accept"); got != "application/json" {
				t.Errorf("Accept = %q, want application/json", got)
			}
			if got := request.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", got)
			}
			if got := request.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
				t.Errorf("X-Requested-With = %q, want XMLHttpRequest", got)
			}
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
			return response(http.StatusOK, `{"status":200,"data":{"instanceUuid":"instance"},"time":1718594177859}`), nil
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
			return response(http.StatusServiceUnavailable, "unavailable"), nil
		})},
	}
	if err := config.SendCommand(context.Background(), "list"); err == nil {
		t.Fatal("SendCommand() error = nil, want HTTP status error")
	}
}

func TestSendCommandReportsAPIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		detail    string
		wantError string
	}{
		{name: "bad request", status: http.StatusBadRequest, detail: "invalid uuid", wantError: "400（请求参数不正确）: invalid uuid"},
		{name: "forbidden", status: http.StatusForbidden, detail: "invalid apikey", wantError: "403（权限不足）: invalid apikey"},
		{name: "server error", status: http.StatusInternalServerError, detail: "internal error", wantError: "500（程序错误）: internal error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			body := fmt.Sprintf(`{"status":%d,"data":%q,"time":1718594177859}`, test.status, test.detail)
			config := Config{
				APIURL: "https://mcsm.example.test/command",
				httpClient: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
					return response(http.StatusOK, body), nil
				})},
			}
			err := config.SendCommand(context.Background(), "list")
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("SendCommand() error = %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestSendCommandRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	config := Config{
		APIURL: "https://mcsm.example.test/command",
		httpClient: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return response(http.StatusOK, "not-json"), nil
		})},
	}
	if err := config.SendCommand(context.Background(), "list"); err == nil {
		t.Fatal("SendCommand() error = nil, want JSON decoding error")
	}
}

func response(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
