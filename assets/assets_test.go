package assets_test

import (
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
