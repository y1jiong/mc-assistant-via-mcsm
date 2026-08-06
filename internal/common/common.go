package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadJSON(filePath string, destination any) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取 %q: %w", filePath, err)
	}
	if err := json.Unmarshal(content, destination); err != nil {
		return fmt.Errorf("解析 %q: %w", filePath, err)
	}
	return nil
}

func SaveJSON(content any, filePath string) error {
	directory := filepath.Dir(filePath)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("创建目录 %q: %w", directory, err)
	}

	jsonContent, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return fmt.Errorf("编码 JSON: %w", err)
	}
	if err := os.WriteFile(filePath, jsonContent, 0o644); err != nil {
		return fmt.Errorf("写入 %q: %w", filePath, err)
	}
	return nil
}
