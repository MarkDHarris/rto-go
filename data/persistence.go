package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var globalDataDir string

// SetDataDir sets the global data directory (called from main before any load/save).
func SetDataDir(dir string) {
	globalDataDir = dir
}

// GetDataDir returns the configured data directory, defaulting to ./config.
func GetDataDir() string {
	if globalDataDir != "" {
		return globalDataDir
	}
	return "./config"
}

// GetFilePath returns the full path for a data file.
func GetFilePath(name string) string {
	return filepath.Join(GetDataDir(), name)
}

// loadFile reads a file, returning nil bytes (not error) if the file doesn't exist.
func loadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return data, nil
}

// LoadJSON deserializes JSON from a file in the global data directory.
// If the file doesn't exist, v is left unchanged (no error).
func LoadJSON(filename string, v interface{}) error {
	return LoadJSONFrom(GetDataDir(), filename, v)
}

// LoadJSONFrom deserializes JSON from a file in the given directory.
func LoadJSONFrom(dir, filename string, v interface{}) error {
	path := filepath.Join(dir, filename)
	data, err := loadFile(path)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}

// SaveJSON serializes v as JSON and writes to a file in the global data directory.
func SaveJSON(filename string, v interface{}) error {
	return SaveJSONTo(GetDataDir(), filename, v)
}

// SaveJSONTo serializes v as JSON and writes to a file in the given directory.
func SaveJSONTo(dir, filename string, v interface{}) error {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directories for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON for %s: %w", filename, err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadYAML deserializes YAML from a file in the global data directory.
// If the file doesn't exist, v is left unchanged (no error).
func LoadYAML(filename string, v interface{}) error {
	return LoadYAMLFrom(GetDataDir(), filename, v)
}

// LoadYAMLFrom deserializes YAML from a file in the given directory.
func LoadYAMLFrom(dir, filename string, v interface{}) error {
	path := filepath.Join(dir, filename)
	data, err := loadFile(path)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return yaml.Unmarshal(data, v)
}

// SaveYAML serializes v as YAML and writes to a file in the global data directory.
func SaveYAML(filename string, v interface{}) error {
	return SaveYAMLTo(GetDataDir(), filename, v)
}

// SaveYAMLTo serializes v as YAML and writes to a file in the given directory.
func SaveYAMLTo(dir, filename string, v interface{}) error {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directories for %s: %w", path, err)
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling YAML for %s: %w", filename, err)
	}
	return os.WriteFile(path, data, 0644)
}
