package dependencies

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/assets"
)

type config struct {
	initialized bool
	url         string
	integrity   string
	crossorigin string
	nonce       string
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
		cfg.url = assets.RuntimeCDNURL
		cfg.integrity = assets.RuntimeCDNIntegrity
		cfg.crossorigin = "anonymous"
	})
}

// WithCDNURL selects an application-owned CDN URL and optional integrity
// value. An empty URL leaves the current source unchanged.
func WithCDNURL(url, integrity string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		cfg.url = url
		cfg.integrity = integrity
		cfg.crossorigin = "anonymous"
	})
}

// WithLocalURL selects an application-owned local runtime path. An empty URL
// leaves the versioned embedded path unchanged.
func WithLocalURL(url string) Option {
	return optionFunc(func(cfg *config) {
		if url == "" {
			return
		}
		cfg.url = url
		cfg.integrity = ""
		cfg.crossorigin = ""
	})
}

func newConfig(options []Option) config {
	cfg := config{initialized: true, url: assets.RuntimeURL}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) attributes() templ.Attributes {
	attributes := templ.Attributes{}
	if cfg.integrity != "" {
		attributes["integrity"] = cfg.integrity
	}
	if cfg.crossorigin != "" {
		attributes["crossorigin"] = cfg.crossorigin
	}
	if cfg.nonce != "" {
		attributes["nonce"] = cfg.nonce
	}
	return attributes
}
