package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mc-assistant-via-mcsm/internal/common"
)

type Team struct {
	TeamName string   `json:"team_name"`
	Members  []string `json:"members"`
}

type Teams struct {
	Teams         []Team
	ID            map[string]struct{}
	TPCoordinates []string
	NoTeam        bool
}

type commandSender interface {
	SendCommand(context.Context, string) error
	Delay(context.Context) error
}

func NewTeams() *Teams {
	return &Teams{
		Teams: make([]Team, 0, 4),
		ID:    make(map[string]struct{}),
	}
}

func (t *Teams) LoadJSONFile(fileName string) error {
	if err := common.LoadJSON(fileName, &t.Teams); err != nil {
		return err
	}
	return nil
}

func (t *Teams) ParseTeamAndMember(teamDirectoryName string) error {
	entries, err := os.ReadDir(teamDirectoryName)
	if err != nil {
		return fmt.Errorf("读取队伍目录 %q: %w", teamDirectoryName, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		teamName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		path := filepath.Join(teamDirectoryName, entry.Name())
		log.Printf("加载队伍 %s (%s)", teamName, entry.Name())
		if err := t.loadTextFile(teamName, path); err != nil {
			return err
		}
	}
	return nil
}

func (t *Teams) loadTextFile(teamName, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取队伍文件 %q: %w", filePath, err)
	}

	team := Team{TeamName: teamName}
	var duplicateErrors []error
	for _, member := range splitLines(string(content)) {
		if member == "" {
			continue
		}
		if _, exists := t.ID[member]; exists {
			duplicateErrors = append(duplicateErrors, fmt.Errorf("检查到重复 ID: %s", member))
			continue
		}
		team.Members = append(team.Members, member)
		t.ID[member] = struct{}{}
	}
	if err := errors.Join(duplicateErrors...); err != nil {
		return err
	}

	t.Teams = append(t.Teams, team)
	return nil
}

func (t *Teams) ExecuteWhiteTeamCommand(ctx context.Context, sender commandSender) error {
	for _, team := range t.Teams {
		if !t.NoTeam {
			if err := sendWithDelay(ctx, sender, "team add "+team.TeamName); err != nil {
				return err
			}
		}
		for _, member := range team.Members {
			if err := sendWithDelay(ctx, sender, "whitelist add "+member); err != nil {
				return err
			}
			if !t.NoTeam {
				if err := sendWithDelay(ctx, sender, "team join "+team.TeamName+" "+member); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Teams) ParseCoordinate(coordinateFile string) error {
	content, err := os.ReadFile(coordinateFile)
	if err != nil {
		return fmt.Errorf("读取坐标文件 %q: %w", coordinateFile, err)
	}

	t.TPCoordinates = t.TPCoordinates[:0]
	for _, coordinate := range splitLines(string(content)) {
		if coordinate != "" {
			t.TPCoordinates = append(t.TPCoordinates, coordinate)
		}
	}
	if len(t.TPCoordinates) == 0 {
		return errors.New("坐标文件中没有可用坐标")
	}
	return nil
}

func (t *Teams) ExecuteTPCommand(ctx context.Context, sender commandSender, teamName string, countPerCoordinate int) error {
	if countPerCoordinate < 1 {
		return errors.New("每个坐标的玩家数必须大于 0")
	}
	if len(t.TPCoordinates) == 0 {
		return errors.New("没有可用的传送坐标")
	}

	for _, team := range t.Teams {
		if team.TeamName != teamName {
			continue
		}
		for index, member := range team.Members {
			coordinateIndex := (index / countPerCoordinate) % len(t.TPCoordinates)
			command := "tp " + member + " " + t.TPCoordinates[coordinateIndex]
			if err := sendWithDelay(ctx, sender, command); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("未找到队伍 %q", teamName)
}

func sendWithDelay(ctx context.Context, sender commandSender, command string) error {
	if err := sender.SendCommand(ctx, command); err != nil {
		return err
	}
	return sender.Delay(ctx)
}

func splitLines(content string) []string {
	return strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
}
