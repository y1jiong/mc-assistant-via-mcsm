package common

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	configFileName = "config.json"
)

type Config struct {
	ApiUrl            string `json:"api_url"`
	ApiKey            string `json:"api_key"`
	NodeId            string `json:"node_id"`
	InstanceId        string `json:"instance_id"`
	DefaultDataFile   string `json:"default_data_file"`
	httpClient        *http.Client
	DelayMilliseconds int `json:"-"`
	delayDuration     time.Duration
}

func (s *Config) InitToFile() (err error) {
	*s = Config{
		ApiUrl:          "http://127.0.0.1:23333/api/protected_instance/command",
		DefaultDataFile: "data.json",
	}
	err = MarshalAndSave(s, configFileName)
	if err != nil {
		return
	}
	log.Println("初始化 " + configFileName + " 完成")
	return
}

func (s *Config) LoadFromFile() (err error) {
	return LoadAndUnmarshal(configFileName, s)
}

func (s *Config) Init(insecure bool) {
	s.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
	s.DelayMilliseconds = 550
	s.delayDuration = time.Duration(s.DelayMilliseconds) * time.Millisecond
}

func (s *Config) SetDelay(milliseconds int) {
	s.DelayMilliseconds = milliseconds
	s.delayDuration = time.Duration(milliseconds) * time.Millisecond
}

func (s *Config) SendCommand(command string) (err error) {
	log.Println(command)
	// 准备请求
	req, err := http.NewRequest("GET", s.ApiUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	q := req.URL.Query()
	q.Add("apikey", s.ApiKey)
	q.Add("daemonId", s.NodeId)
	q.Add("uuid", s.InstanceId)
	q.Add("command", command)
	req.URL.RawQuery = q.Encode()

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(resp.StatusCode) + " " + http.StatusText(resp.StatusCode))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	err = resp.Body.Close()
	return
}

func (s *Config) Delay() {
	time.Sleep(s.delayDuration)
}

func (s *Config) NewTicker() *time.Ticker {
	return time.NewTicker(s.delayDuration)
}
