package common

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	configFileName = "config.json"
)

type Config struct {
	ApiUrl              string `json:"api_url"`
	ApiKey              string `json:"api_key"`
	GID                 string `json:"gid"`
	UID                 string `json:"uid"`
	DefaultDataFileName string `json:"default_data_file_name"`
	httpClient          http.Client
	DelayMilliseconds   int `json:"-"`
	delayDuration       time.Duration
}

func (c *Config) InitToFile() (err error) {
	*c = Config{
		ApiUrl:              "http://127.0.0.1:23333/api/protected_instance/command",
		DefaultDataFileName: "data.json",
	}
	if err != nil {
		return
	}
	err = MarshalAndSave(c, configFileName)
	if err != nil {
		return
	}
	log.Println("初始化 " + configFileName + " 完成")
	return
}

func (c *Config) LoadFromFile() (err error) {
	return LoadAndUnmarshal(configFileName, c)
}

func (c *Config) Init(insecure bool) {
	c.httpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
	c.DelayMilliseconds = 550
	c.delayDuration = time.Duration(c.DelayMilliseconds) * time.Millisecond
}

func (c *Config) SetDelay(milliseconds int) {
	c.DelayMilliseconds = milliseconds
	c.delayDuration = time.Duration(milliseconds) * time.Millisecond
}

func (c *Config) SendCommand(command string) (err error) {
	log.Println(command)
	// 准备请求
	req, err := http.NewRequest("GET", c.ApiUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	q := req.URL.Query()
	q.Add("uuid", c.UID)
	q.Add("remote_uuid", c.GID)
	q.Add("apikey", c.ApiKey)
	q.Add("command", command)
	req.URL.RawQuery = q.Encode()

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("statusCode is " + strconv.Itoa(resp.StatusCode))
	}
	err = resp.Body.Close()
	return
}

func (c *Config) Delay() {
	time.Sleep(c.delayDuration)
}

func (c *Config) NewTicker() *time.Ticker {
	return time.NewTicker(c.delayDuration)
}
