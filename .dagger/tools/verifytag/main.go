package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	githubAPI        = "https://api.github.com"
	assetsRepository = "araihu/assets"
)

type gitObject struct {
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

func main() {
	release := flag.String("release", "", "stable Arai Hu release tag")
	revision := flag.String("revision", "", "expected full release commit SHA")
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fatal(errors.New("GITHUB_TOKEN is required"))
	}
	if err := verifyTag(context.Background(), http.DefaultClient, githubAPI, token, *release, *revision); err != nil {
		fatal(err)
	}
	fmt.Println("Arai Hu release tag matches dispatched revision")
}

func verifyTag(ctx context.Context, client *http.Client, baseURL, token, release, revision string) error {
	objectType, objectSHA, err := fetchObject(
		ctx,
		client,
		baseURL+"/repos/"+assetsRepository+"/git/ref/tags/"+url.PathEscape(release),
		token,
	)
	if err != nil {
		return fmt.Errorf("resolve release tag: %w", err)
	}
	if objectType == "tag" {
		objectType, objectSHA, err = fetchObject(
			ctx,
			client,
			baseURL+"/repos/"+assetsRepository+"/git/tags/"+url.PathEscape(objectSHA),
			token,
		)
		if err != nil {
			return fmt.Errorf("dereference annotated release tag: %w", err)
		}
	}
	if objectType != "commit" {
		return fmt.Errorf("release tag resolves to %q, not commit", objectType)
	}
	if objectSHA != revision {
		return errors.New("release tag commit does not match dispatched revision")
	}
	return nil
}

func fetchObject(ctx context.Context, client *http.Client, endpoint, token string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "goshtoso-charts-dagger")

	response, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned %s", response.Status)
	}

	var object gitObject
	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return "", "", errors.New("decode GitHub tag identity")
	}
	if object.Object.Type == "" || object.Object.SHA == "" {
		return "", "", errors.New("GitHub tag identity is incomplete")
	}
	return object.Object.Type, object.Object.SHA, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
