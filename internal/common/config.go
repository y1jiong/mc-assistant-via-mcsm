package common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	configFileName  = "config.json"
	defaultDelay    = 550 * time.Millisecond
	requestTimeout  = 30 * time.Second
	maxResponseSize = 1 << 20
)

type commandResponse struct {
	Status int             `json:"status"`
	Data   json.RawMessage `json:"data"`
	Time   int64           `json:"time"`
}

type Config struct {
	APIURL          string `json:"api_url"`
	APIKey          string `json:"api_key"`
	NodeID          string `json:"node_id"`
	InstanceID      string `json:"instance_id"`
	DefaultDataFile string `json:"default_data_file"`

	DelayMilliseconds int `json:"-"`
	httpClient        *http.Client
	delay             time.Duration
}

func (c *Config) InitToFile() error {
	*c = Config{
		APIURL:          "http://127.0.0.1:23333/api/protected_instance/command",
		DefaultDataFile: "data.json",
	}
	if err := SaveJSON(c, configFileName); err != nil {
		return err
	}
	log.Println("初始化 " + configFileName + " 完成")
	return nil
}

func (c *Config) LoadFromFile() error {
	return LoadJSON(configFileName, c)
}

func (c *Config) Init(insecure bool) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: insecure, //nolint:gosec // 由显式的 --insecure 选项控制。
	}
	c.httpClient = &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}
	c.SetDelay(int(defaultDelay / time.Millisecond))
}

func (c *Config) SetDelay(milliseconds int) {
	c.DelayMilliseconds = milliseconds
	c.delay = time.Duration(milliseconds) * time.Millisecond
}

func (c *Config) SendCommand(ctx context.Context, command string) error {
	if c.httpClient == nil {
		return errors.New("配置尚未初始化")
	}

	log.Println(command)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL, nil)
	if err != nil {
		return fmt.Errorf("创建命令请求: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	query := req.URL.Query()
	query.Set("apikey", c.APIKey)
	query.Set("daemonId", c.NodeID)
	query.Set("uuid", c.InstanceID)
	query.Set("command", command)
	req.URL.RawQuery = query.Encode()

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送命令 %q: %w", command, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("读取命令 %q 的响应: %w", command, err)
	}
	if len(body) > maxResponseSize {
		return fmt.Errorf("命令 %q 的响应超过 %d 字节", command, maxResponseSize)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("发送命令 %q: MCSM 返回 %s%s", command, response.Status, responseDetail(body))
	}

	var result commandResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析命令 %q 的响应: %w", command, err)
	}
	if result.Status != http.StatusOK {
		return fmt.Errorf(
			"发送命令 %q: MCSM 业务状态为 %d（%s）%s",
			command,
			result.Status,
			commandStatusText(result.Status),
			responseDetail(result.Data),
		)
	}
	return nil
}

func commandStatusText(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "请求参数不正确"
	case http.StatusForbidden:
		return "权限不足"
	case http.StatusInternalServerError:
		return "程序错误"
	default:
		return "未知错误"
	}
}

func responseDetail(content []byte) string {
	detail := strings.TrimSpace(string(content))
	if detail == "" || detail == "null" {
		return ""
	}

	var message string
	if json.Unmarshal([]byte(detail), &message) == nil {
		return ": " + message
	}
	return ": " + detail
}

func (c *Config) Delay(ctx context.Context) error {
	timer := time.NewTimer(c.delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Config) DelayDuration() time.Duration {
	return c.delay
}
