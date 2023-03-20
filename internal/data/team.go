package data

import (
	"errors"
	"log"
	"mc-assistant-via-mcsm/internal/common"
	"os"
	"strings"
)

type TeamFormat struct {
	TeamName string   `json:"team_name"`
	Members  []string `json:"members"`
}

type TeamsFormat struct {
	Teams         []TeamFormat
	ID            map[string]bool
	TpCoordinates []string
	NoTeam        bool
}

func (f *TeamsFormat) Init() {
	f.Teams = make([]TeamFormat, 0, 4)
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

// ExecuteWhiteTeamCommand 目前仅支持 MCSM 9
func (f *TeamsFormat) ExecuteWhiteTeamCommand(c common.Config) (err error) {
	// 拼接最终 API 地址
	for _, v := range f.Teams {
		// 创建队伍
		if !f.NoTeam {
			err = c.SendCommand("team add " + v.TeamName)
			if err != nil {
				return
			}
			c.Delay()
		}
		for _, vv := range v.Members {
			// 加入白名单
			err = c.SendCommand("whitelist add " + vv)
			if err != nil {
				return
			}
			c.Delay()

			// 加入队伍
			if !f.NoTeam {
				err = c.SendCommand("team join " + v.TeamName + " " + vv)
				if err != nil {
					return
				}
				c.Delay()
			}
		}
	}
	return
}

func (f *TeamsFormat) ParseCoordinate(coordinateFile string) (err error) {
	_, err = os.Stat(coordinateFile)
	if err != nil {
		return
	}
	txtContent, err := os.ReadFile(coordinateFile)
	// CRLF to LF
	content := strings.ReplaceAll(string(txtContent), "\r\n", "\n")
	if err != nil {
		return
	}
	for _, v := range strings.Split(content, "\n") {
		// 检查空行
		if v != "" {
			(*f).TpCoordinates = append((*f).TpCoordinates, v)
		}
	}
	return
}

func (f *TeamsFormat) ExecuteTpCommand(c common.Config, tpTeam string, tpCountPerCoordinate int) (err error) {
	maxPosition := len(f.TpCoordinates)
	position := 0
	count := 0
	for _, v := range f.Teams {
		if v.TeamName == tpTeam {
			for _, vv := range v.Members {
				// tp sb. coordinate
				err = c.SendCommand("tp " + vv + " " + f.TpCoordinates[position])
				count++
				if count >= tpCountPerCoordinate {
					position++
					if position >= maxPosition {
						position = 0
					}
					count = 0
				}
				if err != nil {
					return
				}
				c.Delay()
			}
			break
		}
	}
	return
}
