package common

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveAndLoadJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	want := payload{Name: "测试", Count: 2}
	path := filepath.Join(t.TempDir(), "nested", "data.json")

	if err := SaveJSON(want, path); err != nil {
		t.Fatalf("SaveJSON() error = %v", err)
	}
	var got payload
	if err := LoadJSON(path, &got); err != nil {
		t.Fatalf("LoadJSON() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadJSON() = %+v, want %+v", got, want)
	}
}
