package dependencies

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/assets"
)

type config struct {
	initialized bool
	core        scriptConfig
	wordCloud   scriptConfig
	liquid      scriptConfig
	gl          scriptConfig
	brazilMap   scriptConfig
	saoPauloMap scriptConfig
	nonce       string
}

type scriptConfig struct {
	url         string
	integrity   string
	crossorigin string
}

// Option configures the runtime tag emitted by Dependencies.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (option optionFunc) apply(cfg *config) { option(cfg) }

// WithCDN selects the pinned CDN runtime. Local embedded delivery remains the
// default unless this option is supplied explicitly.
func WithCDN() Option {
	return optionFunc(func(cfg *config) {
		cfg.core = scriptConfig{url: assets.RuntimeCDNURL, integrity: assets.RuntimeCDNIntegrity, crossorigin: "anonymous"}
		cfg.wordCloud = scriptConfig{url: assets.WordCloudRuntimeCDNURL, integrity: assets.WordCloudRuntimeCDNIntegrity, crossorigin: "anonymous"}
		cfg.liquid = scriptConfig{url: assets.LiquidRuntimeCDNURL, integrity: assets.LiquidRuntimeCDNIntegrity, crossorigin: "anonymous"}
		cfg.gl = scriptConfig{url: assets.ThreeDRuntimeCDNURL, integrity: assets.ThreeDRuntimeCDNIntegrity, crossorigin: "anonymous"}
		// Derived geography has no executable upstream CDN artifact. Keep pinned,
		// embedded registration scripts local while moving published runtimes to CDN.
	})
}

// WithCDNURL selects an application-owned CDN URL and optional integrity
// value. An empty URL leaves the current source unchanged.
func WithCDNURL(url, integrity string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		cfg.core = scriptConfig{url: url, integrity: integrity, crossorigin: "anonymous"}
	})
}

// WithLocalURL selects an application-owned local runtime path. An empty URL
// leaves the versioned embedded path unchanged.
func WithLocalURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		cfg.core = scriptConfig{url: url}
	})
}

func newConfig(options []Option) config {
	cfg := config{
		initialized: true,
		core:        scriptConfig{url: assets.RuntimeURL},
		wordCloud:   scriptConfig{url: assets.WordCloudRuntimeURL},
		liquid:      scriptConfig{url: assets.LiquidRuntimeURL},
		gl:          scriptConfig{url: assets.ThreeDRuntimeURL},
		brazilMap:   scriptConfig{url: assets.BrazilMapURL},
		saoPauloMap: scriptConfig{url: assets.SaoPauloMapURL},
	}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) scripts() []scriptConfig {
	return []scriptConfig{cfg.core, cfg.wordCloud, cfg.liquid, cfg.gl, cfg.brazilMap, cfg.saoPauloMap}
}

func (cfg config) attributes(script scriptConfig) templ.Attributes {
	attributes := templ.Attributes{}
	if script.integrity != "" {
		attributes["integrity"] = script.integrity
	}
	if script.crossorigin != "" {
		attributes["crossorigin"] = script.crossorigin
	}
	if cfg.nonce != "" {
		attributes["nonce"] = cfg.nonce
	}
	return attributes
}
