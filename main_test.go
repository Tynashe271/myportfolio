package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubPagesDeploymentUpdatesProject(t *testing.T) {
	requests := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/repos/Tynashe271/example/pages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"html_url":"https://example.github.io/example/","public":true}`))
	}))
	defer api.Close()

	lookup := newDeploymentLookup()
	lookup.apiBase = api.URL
	projects := []project{{Title: "Example", URL: "http://localhost:3000", Status: "Built", Repo: "example"}}
	lookup.updateProjects(context.Background(), projects)
	lookup.updateProjects(context.Background(), projects)
	if projects[0].URL != "https://example.github.io/example/" || projects[0].Status != "Live" || !projects[0].Deployed {
		t.Fatalf("project was not updated: %+v", projects[0])
	}
	if requests != 1 {
		t.Fatalf("API requests = %d, want 1", requests)
	}
}

func testApplication(t *testing.T) *application {
	t.Helper()
	app, err := newApplication()
	if err != nil {
		t.Fatal(err)
	}
	app.contactFile = t.TempDir() + "/messages.jsonl"
	return app
}

func TestPortfolio(t *testing.T) {
	recorder := httptest.NewRecorder()
	testApplication(t).routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "Tinashe Nyenyesa — Creative Developer") {
		t.Fatal("homepage does not contain portfolio title")
	}
}

func TestContact(t *testing.T) {
	tests := []struct {
		name, body string
		want       int
	}{
		{"valid", `{"name":"Test User","email":"test@example.com","message":"Hello"}`, http.StatusCreated},
		{"empty", `{"name":"","email":"","message":""}`, http.StatusBadRequest},
		{"invalid email", `{"name":"Test User","email":"not-an-email","message":"Hello"}`, http.StatusBadRequest},
		{"malformed", `{bad json}`, http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/contact/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			testApplication(t).routes().ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}
