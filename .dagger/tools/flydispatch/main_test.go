package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildRequest(t *testing.T) {
	t.Parallel()

	const sha = "0123456789abcdef0123456789abcdef01234567"
	request, err := buildRequest(sha, "12345", "commit-0123456789ab")
	if err != nil {
		t.Fatal(err)
	}
	if request.EventType != "goshtoso-charts-main" {
		t.Fatalf("event type = %q", request.EventType)
	}
	if request.ClientPayload.Ref != sha || request.ClientPayload.SHA != sha {
		t.Fatalf("identity payload = %#v", request.ClientPayload)
	}
	if request.ClientPayload.SourceRepository != "araihu/goshtoso-charts" {
		t.Fatalf("source repository = %q", request.ClientPayload.SourceRepository)
	}
}

func TestDispatch(t *testing.T) {
	t.Parallel()

	const sha = "0123456789abcdef0123456789abcdef01234567"
	payload, err := buildRequest(sha, "12345", "v1.2.3-rc.1+build.7")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		var got dispatchRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if got != payload {
			t.Errorf("payload = %#v, want %#v", got, payload)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := dispatch(context.Background(), server.Client(), server.URL, "test-token", payload); err != nil {
		t.Fatal(err)
	}
}

func TestBuildRequestRejectsInvalidIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sha     string
		runID   string
		version string
	}{
		{name: "short SHA", sha: "0123", runID: "1", version: "v1.2.3"},
		{name: "zero run", sha: "0123456789abcdef0123456789abcdef01234567", runID: "0", version: "v1.2.3"},
		{name: "bad version", sha: "0123456789abcdef0123456789abcdef01234567", runID: "1", version: "latest"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := buildRequest(tt.sha, tt.runID, tt.version); err == nil {
				t.Fatal("invalid identity unexpectedly passed")
			}
		})
	}
}
