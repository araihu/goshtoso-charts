package cicontract

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func workflow(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func argsBlocks(contents string) string {
	var blocks strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(contents))
	inArgs := false
	argsIndent := 0
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if strings.HasPrefix(trimmed, "args:") {
			inArgs = true
			argsIndent = indent
			continue
		}
		if inArgs && trimmed != "" && indent <= argsIndent {
			inArgs = false
		}
		if inArgs {
			blocks.WriteString(trimmed)
			blocks.WriteByte('\n')
		}
	}
	return blocks.String()
}

func TestExternalValuesNeverEnterActionArgs(t *testing.T) {
	tests := []struct {
		workflow string
		payload  string
	}{
		{"araihu-assets.yml", "--payload=.dagger-inputs/assets.json"},
		{"deploy.yml", "--payload=.dagger-inputs/deploy.json"},
	}
	for _, test := range tests {
		t.Run(test.workflow, func(t *testing.T) {
			args := argsBlocks(workflow(t, test.workflow))
			if !strings.Contains(args, test.payload) {
				t.Fatalf("missing literal payload path in action args:\n%s", args)
			}
			for _, forbidden := range []string{
				"${{ github.", "${{ inputs.", "${{ github.event.",
				"${{ env.ASSETS_", "${{ env.RELEASE", "${{ env.REF_", "${{ env.SOURCE_",
			} {
				if strings.Contains(args, forbidden) {
					t.Fatalf("external value %q reached action args:\n%s", forbidden, args)
				}
			}
		})
	}
}

func TestHostingerUsesOnlyEmbeddedExactDagger(t *testing.T) {
	contents := workflow(t, "ci.yml")
	action := "dagger/dagger-for-github@27b130bf0f79a7f6fbbbe0fbca6760dc9bb40a77"
	for _, required := range []string{
		"if: env.HOSTINGER_RUNNER == 'true'",
		"actual_version=$(dagger version",
		`if [ "$actual_version" != "v${DAGGER_VERSION}" ]`,
		"dagger call ci",
		"if: env.HOSTINGER_RUNNER != 'true'",
		action,
		"version: ${{ env.DAGGER_VERSION }}",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("CI workflow missing Hostinger adapter contract %q", required)
		}
	}
	if strings.Contains(contents, "version: latest") {
		t.Fatal("CI workflow may not query or install the latest Dagger version")
	}
	if strings.Count(contents, action) != 1 {
		t.Fatalf("expected exactly one GitHub-hosted Dagger action adapter, got %d", strings.Count(contents, action))
	}
	githubHostedAdapter := "if: env.HOSTINGER_RUNNER != 'true'\n        uses: " + action
	if !strings.Contains(contents, githubHostedAdapter) {
		t.Fatal("Dagger action adapter is not guarded away from Hostinger")
	}
}
