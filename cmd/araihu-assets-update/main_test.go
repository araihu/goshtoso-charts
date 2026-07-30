package main

import (
	"io"
	"strings"
	"testing"
)

func TestRunRequiresReleaseDirectoryAndCompleteIdentity(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "release directory", want: "-release-dir is required"},
		{
			name: "complete identity",
			args: []string{"-release-dir", t.TempDir(), "-assets-repository", "araihu/assets"},
			want: "release identity flags must be provided together",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.args, io.Discard)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("run() error = %v, want %q", err, test.want)
			}
		})
	}
}
