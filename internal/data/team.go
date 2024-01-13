package data

import (
	"errors"
	"log"
	"mc-assistant-via-mcsm/internal/common"
	"os"
	"strings"
)

type Team struct {
	TeamName string   `json:"team_name"`
	Members  []string `json:"members"`
}

type Teams struct {
	Teams         []Team
	Id            map[string]struct{}
	TpCoordinates []string
	NoTeam        bool
}

func (s *Teams) Init() {
	s.Teams = make([]Team, 0, 4)
	s.Id = make(map[string]struct{})
	return
}

func (s *Teams) LoadJsonFile(fileName string) (err error) {
	return common.LoadAndUnmarshal(fileName, &(s.Teams))
}

func (s *Teams) ParseTeamAndMember(teamDirectoryName string) (err error) {
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
			err = s.loadTxtFile(teamName, path)
			if err != nil {
				return
			}
		}
	}
	return
}

func (s *Teams) loadTxtFile(teamName, filePath string) (err error) {
	// 新增队伍数据结构
	t := Team{TeamName: teamName}
	_, err = os.Stat(filePath)
	if err != nil {
		return
	}
	txtContent, err := os.ReadFile(filePath)
	// CRLF to LF
	content := strings.ReplaceAll(string(txtContent), "\r\n", "\n")
	// err text
	errText := strings.Builder{}
	if err != nil {
		return
	}
	for _, v := range strings.Split(content, "\n") {
		// 检查空行
		if v != "" {
			if _, ok := s.Id[v]; ok {
				if errText.Len() > 0 {
					errText.WriteString("\n")
				}
				errText.WriteString("检查到 ID: " + v + " 重复")
				continue
			}
			t.Members = append(t.Members, v)
			s.Id[v] = struct{}{}
		}
	}
	if errText.Len() > 0 {
		return errors.New(errText.String())
	}
	(*s).Teams = append((*s).Teams, t)
	return
}

// ExecuteWhiteTeamCommand 目前仅支持 MCSM 9
func (s *Teams) ExecuteWhiteTeamCommand(c common.Config) (err error) {
	// 拼接最终 API 地址
	for _, v := range s.Teams {
		// 创建队伍
		if !s.NoTeam {
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
			if !s.NoTeam {
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

func (s *Teams) ParseCoordinate(coordinateFile string) (err error) {
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
			(*s).TpCoordinates = append((*s).TpCoordinates, v)
		}
	}
	return
}

func (s *Teams) ExecuteTpCommand(c common.Config, tpTeam string, tpCountPerCoordinate int) (err error) {
	maxPosition := len(s.TpCoordinates)
	position := 0
	count := 0
	for _, v := range s.Teams {
		if v.TeamName == tpTeam {
			for _, vv := range v.Members {
				// tp sb. coordinate
				err = c.SendCommand("tp " + vv + " " + s.TpCoordinates[position])
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
