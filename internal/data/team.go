package data

import (
	"crypto/tls"
	"errors"
	"log"
	"mc-whitelist-team-manager-cli/internal/common"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TeamFormat struct {
	TeamName string   `json:"team_name"`
	Members  []string `json:"members"`
}

type TeamsFormat struct {
	Teams      []TeamFormat
	httpClient http.Client
	ID         map[string]bool
}

func (f *TeamsFormat) Init(insecure bool) {
	f.Teams = make([]TeamFormat, 0, 4)
	f.httpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
	f.ID = make(map[string]bool)
	return
}

func (f *TeamsFormat) LoadJsonFile(fileName string) (err error) {
	return common.LoadAndUnmarshal(fileName, &(f.Teams))
}

func (f *TeamsFormat) ParseTeamAndMember(teamDirectoryName string) (err error) {
	fileInfoList, err := os.ReadDir(teamDirectoryName)
	if err != nil {
		return
	}
	for _, v := range fileInfoList {
		if !v.IsDir() {
			// 获取队伍名称
			teamName := ""
			if strings.Contains(v.Name(), ".") {
				tempSlice := strings.Split(v.Name(), ".")
				tempSlice = tempSlice[0 : len(tempSlice)-1]
				length := len(tempSlice)
				for k, v := range tempSlice {
					teamName += v
					if k+1 < length {
						teamName += "."
					}
				}
			} else {
				teamName = v.Name()
			}

			// 读取队伍成员文件
			path := teamDirectoryName + "/" + v.Name()
			log.Println("加载队伍", teamName, "("+v.Name()+")")
			err = f.loadTxtFile(teamName, path)
			if err != nil {
				return
			}
		}
	}
	return
}

func (f *TeamsFormat) loadTxtFile(teamName, filePath string) (err error) {
	// 新增队伍数据结构
	t := TeamFormat{TeamName: teamName}
	_, err = os.Stat(filePath)
	if err != nil {
		return
	}
	txtContent, err := os.ReadFile(filePath)
	// CRLF to LF
	content := strings.ReplaceAll(string(txtContent), "\r\n", "\n")
	if err != nil {
		return
	}
	for _, v := range strings.Split(content, "\n") {
		// 检查空行
		if v != "" {
			if f.ID[v] {
				return errors.New("检查到 ID: " + v + " 重复")
			}
			t.Members = append(t.Members, v)
			f.ID[v] = true
		}
	}
	(*f).Teams = append((*f).Teams, t)
	return
}

// ExecuteCommand 目前仅支持 MCSM 8
func (f TeamsFormat) ExecuteCommand(apiUrl, apiKey, serverName string) (err error) {
	// 拼接最终 API 地址
	dstUrl := apiUrl + "/?apikey=" + apiKey
	for _, v := range f.Teams {
		// 创建队伍
		err = f.postCommand(dstUrl, serverName, "team add "+v.TeamName)
		if err != nil {
			return
		}
		for _, vv := range v.Members {
			time.Sleep(1100 * time.Millisecond)
			// 加入白名单
			err = f.postCommand(dstUrl, serverName, "whitelist add "+vv)
			if err != nil {
				return
			}
			time.Sleep(1100 * time.Millisecond)
			// 加入队伍
			err = f.postCommand(dstUrl, serverName, "team join "+v.TeamName+" "+vv)
			if err != nil {
				return
			}
		}
		time.Sleep(1100 * time.Millisecond)
	}
	return
}

func (f TeamsFormat) postCommand(apiUrl, serverName, command string) (err error) {
	// 准备请求
	uv := url.Values{}
	uv.Add("name", serverName)
	uv.Add("command", command)
	log.Println(command)
	body := uv.Encode()
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	err = req.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("statusCode is not 200")
	}
	err = resp.Body.Close()
	return
}
