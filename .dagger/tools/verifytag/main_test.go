package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyTag(t *testing.T) {
	t.Parallel()

	const revision = "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name      string
		annotated bool
	}{
		{name: "lightweight"},
		{name: "annotated", annotated: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
					t.Errorf("Authorization = %q", got)
				}
				if tt.annotated && r.URL.Path == "/repos/araihu/assets/git/ref/tags/v1.2.3" {
					fmt.Fprint(w, `{"object":{"type":"tag","sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)
					return
				}
				fmt.Fprintf(w, `{"object":{"type":"commit","sha":%q}}`, revision)
			}))
			defer server.Close()

			if err := verifyTag(context.Background(), server.Client(), server.URL, "test-token", "v1.2.3", revision); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVerifyTagRejectsRevisionMismatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"object":{"type":"commit","sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)
	}))
	defer server.Close()

	err := verifyTag(
		context.Background(),
		server.Client(),
		server.URL,
		"test-token",
		"v1.2.3",
		"0123456789abcdef0123456789abcdef01234567",
	)
	if err == nil {
		t.Fatal("revision mismatch unexpectedly passed")
	}
}
