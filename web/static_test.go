//nolint:testpackage // verifies the package's embedded production assets
package web

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEmbeddedOpenAPISpecParses(t *testing.T) {
	t.Parallel()

	data, err := static.ReadFile("static/spec/spec.yaml")
	if err != nil {
		t.Fatalf("read embedded OpenAPI spec: %v", err)
	}

	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatalf("parse embedded OpenAPI spec: %v", err)
	}
	if document["openapi"] != "3.0.3" {
		t.Fatalf("openapi version = %v", document["openapi"])
	}
}

func TestEmbeddedInterfaceHasDarkDefaultAndControls(t *testing.T) {
	t.Parallel()

	data, err := static.ReadFile("static/templates/index.html")
	if err != nil {
		t.Fatalf("read embedded interface: %v", err)
	}
	page := string(data)

	for _, required := range []string{
		`<html lang="en" data-theme="dark">`,
		`id="theme-toggle"`,
		`value="auto" selected`,
		`id="reset-form"`,
		`pinharbor-logo.png`,
	} {
		if !strings.Contains(page, required) {
			t.Errorf("interface is missing %q", required)
		}
	}

	for _, asset := range []string{"static/css/monitor.css", "static/js/app.js", "static/img/pinharbor-logo.png"} {
		if _, err := static.ReadFile(asset); err != nil {
			t.Errorf("missing embedded asset %s: %v", asset, err)
		}
	}

	jobRow, err := static.ReadFile("static/templates/job_row.html")
	if err != nil {
		t.Fatalf("read embedded job row: %v", err)
	}
	for _, required := range []string{`hx-post="/stop?id=`, "Stop &amp; keep partial"} {
		if !strings.Contains(string(jobRow), required) {
			t.Errorf("job row is missing %q", required)
		}
	}

	styles, err := static.ReadFile("static/css/monitor.css")
	if err != nil {
		t.Fatalf("read embedded theme: %v", err)
	}
	for _, required := range []string{"[data-theme=\"light\"]", ".status-stopped", ".stop-button"} {
		if !strings.Contains(string(styles), required) {
			t.Errorf("theme is missing %q", required)
		}
	}
}
