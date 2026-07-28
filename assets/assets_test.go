package assets_test

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/assets"
)

func TestHandlerServesVersionedRuntimeAtDefaultMount(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.RuntimeURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.RuntimeURL, recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `version="5.4.3"`) {
		t.Fatal("embedded runtime does not report pinned version 5.4.3")
	}
}

func TestHandlerDoesNotServeUnversionedRuntimeAlias(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/charts/assets/echarts.min.js", nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GET unversioned runtime status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestHandlerServesVersionedChartControlRuntime(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.ControlRuntimeURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.ControlRuntimeURL, recorder.Code, http.StatusOK)
	}
	for _, want := range []string{"requestFullscreen", "getDataURL", "goshtoso-charts:resize"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("control runtime missing %q", want)
		}
	}
	if strings.Contains(recorder.Body.String(), "chart.resize()") {
		t.Fatal("renderer-neutral controls runtime resized a private chart engine directly")
	}
}

func TestHandlerKeepsPreviousChartControlRuntimeAvailable(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/charts/assets/js/controls/1/controls.js", nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET previous controls runtime status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "goshtoso-charts:resize") {
		t.Fatal("previous controls runtime is not intact")
	}
}

func TestHandlerServesVersionedWordCloudRuntime(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.WordCloudRuntimeURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.WordCloudRuntimeURL, recorder.Code, http.StatusOK)
	}
	for _, want := range []string{"wordCloud", "layoutAnimation", "drawOutOfBound"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("word-cloud runtime missing %q", want)
		}
	}
}

func TestHandlerServesVersionedLiquidRuntime(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.LiquidRuntimeURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.LiquidRuntimeURL, recorder.Code, http.StatusOK)
	}
	for _, want := range []string{"liquidFill", "waveAnimation", "backgroundStyle"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("liquid runtime missing %q", want)
		}
	}
}

func TestHandlerServesPinnedThreeDRuntimeAndLicense(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		url, contains, sha256 string
	}{
		{assets.ThreeDRuntimeURL, "scatter3D", "bfba1b87b8c3c06e5c7ed7741002586c747b00e4efdaa92077d15c2dc721bda0"},
		{assets.ThreeDLicenseURL, "BSD 3-Clause License", "55ea01207028f76d844678511f29fc800f8a1e67a8a0fe80470128677847ad32"},
	} {
		recorder := httptest.NewRecorder()
		assets.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.url, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", test.url, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), test.contains) {
			t.Errorf("GET %s missing %q", test.url, test.contains)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(recorder.Body.Bytes())); got != test.sha256 {
			t.Errorf("GET %s SHA-256 = %s, want %s", test.url, got, test.sha256)
		}
	}
}

func TestHandlerServesLiquidRuntimeLicense(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.LiquidLicenseURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.LiquidLicenseURL, recorder.Code, http.StatusOK)
	}
	for _, want := range []string{"BSD 3-Clause License", "Copyright (c) 2020, Baidu Inc.", "Redistribution and use"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("liquid runtime license missing %q", want)
		}
	}
}

func TestHandlerServesPinnedMapResources(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ url, registration, sha256 string }{
		{assets.ChinaMapURL, `registerMap("china"`, "146a69f110aca347228447319216ad665fbf6a57d81c73ddc911c1167aa39249"},
		{assets.GuangdongMapURL, `registerMap("广东"`, "ca870acf1f735d4b8fda33bb41c0a2804c320ee0885a772e428cfdc4d66f4757"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, test.url, nil)
		assets.Handler().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", test.url, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), test.registration) {
			t.Errorf("map resource %s missing %q", test.url, test.registration)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(recorder.Body.Bytes())); got != test.sha256 {
			t.Errorf("map resource %s SHA-256 = %s, want %s", test.url, got, test.sha256)
		}
	}
}
