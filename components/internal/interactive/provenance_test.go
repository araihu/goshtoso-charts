package interactive

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestSharedTemplateRuntimeProvenance(t *testing.T) {
	t.Parallel()

	for path, want := range map[string]string{
		"interactive.templ":    "96137171bb2e6cb69372a59596963d0f4e7e6f87f3079863ccc92f3f59f8680a",
		"interactive_templ.go": "9e83e969108d203fe38fce502a21fed7c0f85d150cd8978d4ca76e596d278808",
		"live_runtime.go":      "52feab4a14c172ffe212fb95b98e0363293ad0ad253ae60e8caea22caf7f2a4b",
		"theme_runtime.go":     "07607b72118cf2e2e1cc71d81ec3c64789dd2f053ff0f9282a2e41f92cbf24ae",
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read shared provenance file %s: %v", path, err)
		}
		digest := sha256.Sum256(contents)
		if got := hex.EncodeToString(digest[:]); got != want {
			t.Errorf("shared provenance file %s SHA-256 = %s, want %s", path, got, want)
		}
	}
}
