package assets_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/assets"
)

type mapFeature struct {
	Geometry   json.RawMessage `json:"geometry"`
	Properties struct {
		Name, UF, IBGECode string
	} `json:"properties"`
}

func TestHandlerServesVersionedRuntimeAtDefaultMount(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, assets.RuntimeURL, nil)
	assets.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assets.RuntimeURL, recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `version="5.6.0"`) {
		t.Fatal("embedded runtime does not report pinned version 5.6.0")
	}
	const wantSHA256 = "bf4a223524e40b77c304bec67e1222cf551f14880cf42c69dc046558e11c07b1"
	if got := fmt.Sprintf("%x", sha256.Sum256(recorder.Body.Bytes())); got != wantSHA256 {
		t.Fatalf("GET %s SHA-256 = %s, want %s", assets.RuntimeURL, got, wantSHA256)
	}
}

func TestHandlerServesPinnedRuntimeLegalFiles(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		url, contains, sha256 string
	}{
		{assets.RuntimeLicenseURL, "Apache License", "634293835b43a6dd2094fa39182a3d9a6b9ca43b7fdb9ac354e8037af2a3093a"},
		{assets.RuntimeNoticeURL, "Apache ECharts", "4dd56fd5a0ac348fb8cf5dc46d8ce0a7090fb1856ce39c5baa90e13f9ae356c1"},
		{assets.RuntimeD3LicenseURL, "Copyright 2010-2016 Mike Bostock", "e1211892da0b0e0585b7aebe8f98c1274fba15bafe47fa1f4ee8a7a502c06304"},
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
	const wantSHA256 = "7747ad340743c428ff733c7f38e112be828fea268c75fe6ee5c8ca273af106fa"
	if got := fmt.Sprintf("%x", sha256.Sum256(recorder.Body.Bytes())); got != wantSHA256 {
		t.Fatalf("GET %s SHA-256 = %s, want immutable v4 %s", assets.ControlRuntimeURL, got, wantSHA256)
	}
	for _, want := range []string{
		"requestFullscreen", "goshtoso-charts:resize", "goshtoso-charts:export-request", "expandFromMenu", "toggleFullscreen",
		"goshtoso-charts:set-wrapper-mode", "goshtoso-charts:wrapper-mode-change",
		"MutationObserver", "htmx:load", "htmx:afterSwap", "goshtosoChartWrapperMode",
	} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("control runtime missing %q", want)
		}
	}
	for _, engineReference := range []string{"window.echarts", "getInstanceByDom", "getDataURL", "chart.resize()"} {
		if strings.Contains(recorder.Body.String(), engineReference) {
			t.Fatalf("renderer-neutral controls runtime references private chart engine through %q", engineReference)
		}
	}
	for _, unwanted := range []string{"toggleCollapse", "goshtoso-charts-controls-constrained", "updateOverflowVisibility"} {
		if strings.Contains(recorder.Body.String(), unwanted) {
			t.Fatalf("controls runtime retained redundant behavior %q", unwanted)
		}
	}
}

func TestHandlerKeepsPreviousChartControlRuntimesAvailable(t *testing.T) {
	t.Parallel()

	for _, runtime := range []struct{ version, sha256 string }{
		{"1", "ccb5c4c11ab1078549ec02a339f1fc4afdaea747b8d1379ad1dac25d1eb47c5b"},
		{"2", "562c321b5e51c153ba7f6889cce52e15c9fd60f4c0ca70430acb1126708d507d"},
		{"3", "90b1369603f3c77c364e41e5d74a83ff44e8651a9ad8c97e015b3769617de781"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/charts/assets/js/controls/"+runtime.version+"/controls.js", nil)
		assets.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("GET controls v%s runtime status = %d, want %d", runtime.version, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), "goshtoso-charts:resize") {
			t.Errorf("controls v%s runtime is not intact", runtime.version)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(recorder.Body.Bytes())); got != runtime.sha256 {
			t.Errorf("controls v%s SHA-256 = %s, want immutable %s", runtime.version, got, runtime.sha256)
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
		{assets.BrazilMapURL, `registerMap("brazil"`, "1b3719c82f6e2278a3e6ea8b7fc2e195460ee6a7de1546d0a8e05e6d0174bb3d"},
		{assets.SaoPauloMapURL, `registerMap("brazil-sao-paulo"`, "657dee960c4c4d991f5b0e6d59681152d5e2b9c48091e5094085a666c97ff317"},
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

func TestHandlerServesBrazilGeometryLicense(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	assets.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, assets.BrazilMapLicenseURL, nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "Creative Commons Attribution 4.0 International") || !strings.Contains(recorder.Body.String(), "IBGE") {
		t.Fatalf("Brazil geometry license status/content = %d/%q", recorder.Code, recorder.Body.String())
	}
}

func TestBrazilGeometryPreservesExactStateIdentityAndSaoPauloMunicipalities(t *testing.T) {
	t.Parallel()
	type collection struct {
		Features []mapFeature `json:"features"`
	}
	read := func(url, mapName string) collection {
		recorder := httptest.NewRecorder()
		assets.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, url, nil))
		body := recorder.Body.String()
		prefix := `registerMap("` + mapName + `",`
		start := strings.Index(body, prefix)
		end := strings.LastIndex(body, `);})(globalThis);`)
		if start < 0 || end < 0 {
			t.Fatalf("%s registration wrapper missing", url)
		}
		var result collection
		if err := json.Unmarshal([]byte(body[start+len(prefix):end]), &result); err != nil {
			t.Fatalf("decode %s: %v", url, err)
		}
		return result
	}

	brazil := read(assets.BrazilMapURL, "brazil")
	type identity struct{ uf, ibge string }
	want := map[string]identity{
		"Rondônia": {"RO", "11"}, "Acre": {"AC", "12"}, "Amazonas": {"AM", "13"}, "Roraima": {"RR", "14"}, "Pará": {"PA", "15"}, "Amapá": {"AP", "16"}, "Tocantins": {"TO", "17"},
		"Maranhão": {"MA", "21"}, "Piauí": {"PI", "22"}, "Ceará": {"CE", "23"}, "Rio Grande do Norte": {"RN", "24"}, "Paraíba": {"PB", "25"}, "Pernambuco": {"PE", "26"}, "Alagoas": {"AL", "27"}, "Sergipe": {"SE", "28"}, "Bahia": {"BA", "29"},
		"Minas Gerais": {"MG", "31"}, "Espírito Santo": {"ES", "32"}, "Rio de Janeiro": {"RJ", "33"}, "São Paulo": {"SP", "35"}, "Paraná": {"PR", "41"}, "Santa Catarina": {"SC", "42"}, "Rio Grande do Sul": {"RS", "43"},
		"Mato Grosso do Sul": {"MS", "50"}, "Mato Grosso": {"MT", "51"}, "Goiás": {"GO", "52"}, "Distrito Federal": {"DF", "53"},
	}
	if len(brazil.Features) != len(want) {
		t.Fatalf("Brazil geometry features = %d, want %d", len(brazil.Features), len(want))
	}
	for _, feature := range brazil.Features {
		if expected, ok := want[feature.Properties.Name]; !ok || expected.uf != feature.Properties.UF || expected.ibge != feature.Properties.IBGECode {
			t.Errorf("Brazil geometry identity = %#v", feature.Properties)
		}
		delete(want, feature.Properties.Name)
	}
	if len(want) != 0 {
		t.Fatalf("Brazil geometry missing states: %#v", want)
	}
	assertGeometryBounds(t, "Brazil", brazil.Features, [4]float64{-74.1, -33.8, -34.7, 5.3})

	saoPaulo := read(assets.SaoPauloMapURL, "brazil-sao-paulo")
	if len(saoPaulo.Features) != 645 {
		t.Fatalf("São Paulo municipality features = %d, want 645", len(saoPaulo.Features))
	}
	municipalities := map[string]bool{}
	for _, feature := range saoPaulo.Features {
		municipalities[feature.Properties.Name] = true
	}
	for _, name := range []string{"São Paulo", "Campinas", "Ribeirão Preto"} {
		if !municipalities[name] {
			t.Errorf("São Paulo geometry missing municipality %q", name)
		}
	}
	assertGeometryBounds(t, "São Paulo", saoPaulo.Features, [4]float64{-53.2, -25.4, -44.0, -19.7})
}

func assertGeometryBounds(t *testing.T, name string, features []mapFeature, limits [4]float64) {
	t.Helper()
	bounds := [4]float64{180, 90, -180, -90}
	var visit func(any)
	visit = func(value any) {
		switch value := value.(type) {
		case []any:
			if len(value) == 2 {
				longitude, longitudeOK := value[0].(float64)
				latitude, latitudeOK := value[1].(float64)
				if longitudeOK && latitudeOK {
					bounds[0] = min(bounds[0], longitude)
					bounds[1] = min(bounds[1], latitude)
					bounds[2] = max(bounds[2], longitude)
					bounds[3] = max(bounds[3], latitude)
					return
				}
			}
			for _, child := range value {
				visit(child)
			}
		}
	}
	for _, feature := range features {
		var geometry struct {
			Coordinates any `json:"coordinates"`
		}
		if err := json.Unmarshal(feature.Geometry, &geometry); err != nil {
			t.Fatalf("decode %s geometry: %v", name, err)
		}
		visit(geometry.Coordinates)
	}
	if bounds[0] < limits[0] || bounds[1] < limits[1] || bounds[2] > limits[2] || bounds[3] > limits[3] {
		t.Fatalf("%s geometry bounds = %v, outside %v", name, bounds, limits)
	}
}
