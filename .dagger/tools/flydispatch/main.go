package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

const dispatchEndpoint = "https://api.github.com/repos/araihu/fly-deploy/dispatches"

var (
	shaPattern     = regexp.MustCompile(`^[0-9a-f]{40}$`)
	runIDPattern   = regexp.MustCompile(`^[1-9][0-9]*$`)
	versionPattern = regexp.MustCompile(`^(commit-[0-9a-f]{12}|v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?)$`)
)

type dispatchRequest struct {
	EventType     string          `json:"event_type"`
	ClientPayload dispatchPayload `json:"client_payload"`
}

type dispatchPayload struct {
	Ref              string `json:"goshtoso_charts_ref"`
	SHA              string `json:"goshtoso_charts_sha"`
	Version          string `json:"goshtoso_charts_version"`
	RunID            string `json:"goshtoso_charts_run_id"`
	SourceRepository string `json:"source_repository"`
}

func main() {
	sourceSHA := flag.String("source-sha", "", "full Goshtoso Charts source SHA")
	sourceRunID := flag.String("source-run-id", "", "GitHub source run ID")
	version := flag.String("version", "", "validated deployment version")
	flag.Parse()

	token := os.Getenv("FLY_DEPLOY_DISPATCH_TOKEN")
	if token == "" {
		fatal(errors.New("FLY_DEPLOY_DISPATCH_TOKEN is required"))
	}
	payload, err := buildRequest(*sourceSHA, *sourceRunID, *version)
	if err != nil {
		fatal(err)
	}
	if err := dispatch(context.Background(), http.DefaultClient, dispatchEndpoint, token, payload); err != nil {
		fatal(err)
	}
	fmt.Println("Fly deployment dispatch accepted")
}

func buildRequest(sourceSHA, sourceRunID, version string) (dispatchRequest, error) {
	if !shaPattern.MatchString(sourceSHA) {
		return dispatchRequest{}, errors.New("source SHA must be a full lowercase Git SHA-1")
	}
	if !runIDPattern.MatchString(sourceRunID) {
		return dispatchRequest{}, errors.New("source run ID must be a positive decimal GitHub run ID")
	}
	if !versionPattern.MatchString(version) {
		return dispatchRequest{}, errors.New("deployment version is invalid")
	}
	return dispatchRequest{
		EventType: "goshtoso-charts-main",
		ClientPayload: dispatchPayload{
			Ref:              sourceSHA,
			SHA:              sourceSHA,
			Version:          version,
			RunID:            sourceRunID,
			SourceRepository: "araihu/goshtoso-charts",
		},
	}, nil
}

func dispatch(ctx context.Context, client *http.Client, endpoint, token string, payload dispatchRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "goshtoso-charts-dagger")

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("GitHub dispatch API returned %s", response.Status)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
