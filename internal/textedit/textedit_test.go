package textedit

import (
	"errors"
	"testing"
)

func TestSetJSONTopLevelString(t *testing.T) {
	in := `{
  "name": "lib",
  "engines": { "version": "keep-me" },
  "deps": [{"version": "1"}],
  "version":   "1.2.3",
  "private": false
}
`
	out, err := SetJSONTopLevelString([]byte(in), "version", "1.3.0")
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "name": "lib",
  "engines": { "version": "keep-me" },
  "deps": [{"version": "1"}],
  "version":   "1.3.0",
  "private": false
}
`
	if string(out) != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
	if _, err := SetJSONTopLevelString([]byte(`{"name":"x","a":{"version":"1"}}`), "version", "2"); !errors.Is(err, ErrNotFound) {
		t.Errorf("nested-only key: err = %v", err)
	}
}

func TestSetTOMLString(t *testing.T) {
	in := "# crate\n[package]\nname = \"demo\"\nversion = \"0.1.0\" # bumped by sealtag\n\n[dependencies]\nserde = { version = \"1\" }\nversion = \"9\"\n"
	out, err := SetTOMLString([]byte(in), "package", "version", "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	want := "# crate\n[package]\nname = \"demo\"\nversion = \"0.2.0\" # bumped by sealtag\n\n[dependencies]\nserde = { version = \"1\" }\nversion = \"9\"\n"
	if string(out) != want {
		t.Errorf("got:\n%s", out)
	}
	ws := "[workspace]\nmembers = [\"a\"]\n\r\n[workspace.package]\r\nversion = '1.0.0'\r\n"
	out, err = SetTOMLString([]byte(ws), "workspace.package", "version", "1.1.0")
	if err != nil || string(out) != "[workspace]\nmembers = [\"a\"]\n\r\n[workspace.package]\r\nversion = '1.1.0'\r\n" {
		t.Errorf("workspace: %q %v", out, err)
	}
	if _, err := SetTOMLString([]byte("[package]\nname = \"x\"\n"), "package", "version", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing: err = %v", err)
	}
}

func TestPythonVersion(t *testing.T) {
	src := []byte("\"\"\"Docs.\"\"\"\n__version__ = '0.3.0'\n")
	if v, ok := PythonVersion(src); !ok || v != "0.3.0" {
		t.Errorf("read: %q %v", v, ok)
	}
	out, err := SetPythonVersion(src, "0.4.0")
	if err != nil || string(out) != "\"\"\"Docs.\"\"\"\n__version__ = '0.4.0'\n" {
		t.Errorf("set: %q %v", out, err)
	}
}
