package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDataDirDefault(t *testing.T) {
	old := globalDataDir
	globalDataDir = ""
	defer func() { globalDataDir = old }()

	if GetDataDir() != "./config" {
		t.Errorf("expected ./config, got %s", GetDataDir())
	}
}

func TestSetAndGetDataDir(t *testing.T) {
	old := globalDataDir
	defer func() { globalDataDir = old }()

	SetDataDir("/tmp/test")
	if GetDataDir() != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %s", GetDataDir())
	}
}

func TestGetFilePath(t *testing.T) {
	old := globalDataDir
	defer func() { globalDataDir = old }()

	SetDataDir("/tmp/test")
	got := GetFilePath("foo.json")
	want := "/tmp/test/foo.json"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestLoadJSONMissingFile(t *testing.T) {
	dir := t.TempDir()
	type S struct{ Name string }
	var s S
	// Should not return error for missing file
	if err := LoadJSONFrom(dir, "missing.json", &s); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != "" {
		t.Error("expected unchanged zero value")
	}
}

func TestSaveAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	type S struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	orig := S{Name: "Alice", Age: 30}
	if err := SaveJSONTo(dir, "test.json", &orig); err != nil {
		t.Fatalf("save error: %v", err)
	}
	var loaded S
	if err := LoadJSONFrom(dir, "test.json", &loaded); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded != orig {
		t.Errorf("expected %+v, got %+v", orig, loaded)
	}
}

func TestSaveJSONCreatesNestedDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	type S struct{ X int }
	s := S{X: 42}
	if err := SaveJSONTo(nested, "test.json", &s); err != nil {
		t.Fatalf("save error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "test.json")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestLoadYAMLMissingFile(t *testing.T) {
	dir := t.TempDir()
	type S struct{ Name string }
	var s S
	if err := LoadYAMLFrom(dir, "missing.yaml", &s); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveAndLoadYAML(t *testing.T) {
	dir := t.TempDir()
	type S struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	orig := S{Name: "Bob", Age: 25}
	if err := SaveYAMLTo(dir, "test.yaml", &orig); err != nil {
		t.Fatalf("save error: %v", err)
	}
	var loaded S
	if err := LoadYAMLFrom(dir, "test.yaml", &loaded); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded != orig {
		t.Errorf("expected %+v, got %+v", orig, loaded)
	}
}

func TestSaveYAMLCreatesNestedDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "x", "y")
	type S struct{ V string }
	s := S{V: "hello"}
	if err := SaveYAMLTo(nested, "test.yaml", &s); err != nil {
		t.Fatalf("save error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "test.yaml")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
