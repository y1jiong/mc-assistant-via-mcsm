package data

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type recordingSender struct {
	commands []string
}

func (s *recordingSender) SendCommand(_ context.Context, command string) error {
	s.commands = append(s.commands, command)
	return nil
}

func (*recordingSender) Delay(context.Context) error {
	return nil
}

func TestParseTeamAndMember(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "red.team.txt"), []byte("Alice\r\nBob\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	teams := NewTeams()
	if err := teams.ParseTeamAndMember(directory); err != nil {
		t.Fatalf("ParseTeamAndMember() error = %v", err)
	}
	want := []Team{{TeamName: "red.team", Members: []string{"Alice", "Bob"}}}
	if !reflect.DeepEqual(teams.Teams, want) {
		t.Fatalf("ParseTeamAndMember() = %+v, want %+v", teams.Teams, want)
	}
}

func TestParseTeamAndMemberRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "red.txt"), []byte("Alice\nAlice\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := NewTeams().ParseTeamAndMember(directory)
	if err == nil || !strings.Contains(err.Error(), "重复 ID: Alice") {
		t.Fatalf("ParseTeamAndMember() error = %v, want duplicate ID error", err)
	}
}

func TestParseCoordinateRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "coordinates.txt")
	if err := os.WriteFile(path, []byte("\r\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := NewTeams().ParseCoordinate(path); err == nil {
		t.Fatal("ParseCoordinate() error = nil, want empty coordinate error")
	}
}

func TestExecuteWhiteTeamCommandPreservesCommandOrder(t *testing.T) {
	t.Parallel()

	teams := NewTeams()
	teams.Teams = []Team{{TeamName: "red", Members: []string{"Alice", "Bob"}}}
	sender := &recordingSender{}

	if err := teams.ExecuteWhiteTeamCommand(context.Background(), sender); err != nil {
		t.Fatalf("ExecuteWhiteTeamCommand() error = %v", err)
	}
	want := []string{
		"team add red",
		"whitelist add Alice",
		"team join red Alice",
		"whitelist add Bob",
		"team join red Bob",
	}
	if !reflect.DeepEqual(sender.commands, want) {
		t.Fatalf("commands = %q, want %q", sender.commands, want)
	}
}

func TestExecuteTPCommandRotatesCoordinates(t *testing.T) {
	t.Parallel()

	teams := NewTeams()
	teams.Teams = []Team{{TeamName: "red", Members: []string{"A", "B", "C", "D", "E"}}}
	teams.TPCoordinates = []string{"1 2 3", "4 5 6"}
	sender := &recordingSender{}

	if err := teams.ExecuteTPCommand(context.Background(), sender, "red", 2); err != nil {
		t.Fatalf("ExecuteTPCommand() error = %v", err)
	}
	want := []string{
		"tp A 1 2 3",
		"tp B 1 2 3",
		"tp C 4 5 6",
		"tp D 4 5 6",
		"tp E 1 2 3",
	}
	if !reflect.DeepEqual(sender.commands, want) {
		t.Fatalf("commands = %q, want %q", sender.commands, want)
	}
}
