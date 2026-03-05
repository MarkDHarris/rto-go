package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
// String values (but not mapping keys) are always double-quoted so the output
// matches the hand-written style used for holidays, vacations, and settings.
func SaveYAMLTo(dir, filename string, v interface{}) error {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directories for %s: %w", path, err)
	}

	var node yaml.Node
	if err := node.Encode(v); err != nil {
		return fmt.Errorf("encoding YAML for %s: %w", filename, err)
	}
	quoteStringValues(&node)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&node); err != nil {
		return fmt.Errorf("marshaling YAML for %s: %w", filename, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("finalizing YAML for %s: %w", filename, err)
	}

	return os.WriteFile(path, dedentSequences(buf.Bytes(), 2), 0644)
}

// dedentSequences removes the extra indentation that yaml.Encoder adds to
// sequence items, so that `- ` sits at the same column as its parent key
// (style #1) rather than indented under it (style #2).
func dedentSequences(data []byte, indentSize int) []byte {
	lines := strings.Split(string(data), "\n")
	result := make([]string, 0, len(lines))
	inSeq := false
	seqParentIndent := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			result = append(result, line)
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if inSeq && indent <= seqParentIndent {
			inSeq = false
		}

		if inSeq && indent >= indentSize {
			line = line[indentSize:]
		}

		result = append(result, line)

		if strings.HasSuffix(trimmed, ":") && !inSeq {
			for j := i + 1; j < len(lines); j++ {
				nextTrimmed := strings.TrimSpace(lines[j])
				if nextTrimmed == "" {
					continue
				}
				if strings.HasPrefix(nextTrimmed, "- ") {
					inSeq = true
					seqParentIndent = indent
				}
				break
			}
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// quoteStringValues walks a yaml.Node tree and sets DoubleQuotedStyle on
// every string scalar that appears as a mapping value or sequence item,
// leaving mapping keys unquoted.
func quoteStringValues(node *yaml.Node) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			quoteStringValues(child)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			val := node.Content[i+1]
			if val.Kind == yaml.ScalarNode && val.Tag == "!!str" {
				val.Style = yaml.DoubleQuotedStyle
			} else {
				quoteStringValues(val)
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if child.Kind == yaml.ScalarNode && child.Tag == "!!str" {
				child.Style = yaml.DoubleQuotedStyle
			} else {
				quoteStringValues(child)
			}
		}
	}
}
