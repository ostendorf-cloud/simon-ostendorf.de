package config

import (
	"log/slog"
	"strings"
	"testing"
)

// env builds a getenv function over a fixed map, the shape Load expects. A
// key absent from the map behaves exactly like an unset environment
// variable: envconf's own lookup returns "".
func env(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func TestLoadAppliesDefaultsWhenEverythingIsUnset(t *testing.T) {
	t.Parallel()

	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.AdminAddr != ":8081" {
		t.Errorf("AdminAddr = %q, want %q", cfg.AdminAddr, ":8081")
	}
	if cfg.SiteURL.String() != "https://simon-ostendorf.de" {
		t.Errorf("SiteURL = %q, want %q", cfg.SiteURL, "https://simon-ostendorf.de")
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
}

func TestLoadOverridesEachVariable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		env   map[string]string
		check func(t *testing.T, cfg Config)
	}{
		{
			name: "ADDR",
			env:  map[string]string{"ADDR": ":9090"},
			check: func(t *testing.T, cfg Config) {
				if cfg.Addr != ":9090" {
					t.Errorf("Addr = %q, want %q", cfg.Addr, ":9090")
				}
			},
		},
		{
			name: "ADMIN_ADDR",
			env:  map[string]string{"ADMIN_ADDR": ":9091"},
			check: func(t *testing.T, cfg Config) {
				if cfg.AdminAddr != ":9091" {
					t.Errorf("AdminAddr = %q, want %q", cfg.AdminAddr, ":9091")
				}
			},
		},
		{
			name: "SITE_URL",
			env:  map[string]string{"SITE_URL": "https://example.test"},
			check: func(t *testing.T, cfg Config) {
				if cfg.SiteURL.String() != "https://example.test" {
					t.Errorf("SiteURL = %q, want %q", cfg.SiteURL, "https://example.test")
				}
			},
		},
		{
			name: "LOG_LEVEL",
			env:  map[string]string{"LOG_LEVEL": "debug"},
			check: func(t *testing.T, cfg Config) {
				if cfg.LogLevel != slog.LevelDebug {
					t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelDebug)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := Load(env(tc.env))
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			tc.check(t, cfg)
		})
	}
}

func TestLoadUmamiAndProxies(t *testing.T) {
	t.Parallel()

	const id = "123e4567-e89b-42d3-a456-426614174000"

	t.Run("off by default", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load(env(nil))
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Umami.Enabled() {
			t.Error("Umami enabled with nothing set")
		}
		if len(cfg.TrustedProxies) != 0 {
			t.Errorf("TrustedProxies = %v, want none", cfg.TrustedProxies)
		}
	})

	t.Run("both set", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load(env(map[string]string{
			"UMAMI_URL": "https://umami.example.test", "UMAMI_WEBSITE_ID": id,
			"TRUSTED_PROXIES": "10.0.0.0/8, 192.0.2.1",
		}))
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.Umami.Enabled() {
			t.Error("Umami not enabled with both set")
		}
		if len(cfg.TrustedProxies) != 2 {
			t.Errorf("TrustedProxies = %v, want two", cfg.TrustedProxies)
		}
	})

	for name, values := range map[string]map[string]string{
		"only the URL":        {"UMAMI_URL": "https://umami.example.test"},
		"only the website id": {"UMAMI_WEBSITE_ID": id},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(env(values))
			if err == nil || !strings.Contains(err.Error(), "UMAMI_") {
				t.Fatalf("Load() error = %v, want a startup error naming UMAMI_", err)
			}
			if strings.Contains(err.Error(), "umami.example.test") {
				t.Errorf("error repeats the URL: %v", err)
			}
		})
	}

	t.Run("invalid proxy list", func(t *testing.T) {
		t.Parallel()
		if _, err := Load(env(map[string]string{"TRUSTED_PROXIES": "not-a-cidr"})); err == nil {
			t.Fatal("Load() error = nil, want an error for TRUSTED_PROXIES")
		}
	})
}

func TestLoadRejectsAnInvalidLogLevel(t *testing.T) {
	t.Parallel()

	_, err := Load(env(map[string]string{"LOG_LEVEL": "verbose"}))
	if err == nil {
		t.Fatal("Load() error = nil, want an error for an invalid LOG_LEVEL")
	}
	if !strings.Contains(err.Error(), "config:") {
		t.Errorf("Load() error = %q, want it wrapped with \"config:\"", err.Error())
	}
	if !strings.Contains(err.Error(), "verbose") {
		t.Errorf("Load() error = %q, want it to name the invalid value", err.Error())
	}
}

func TestLoadRejectsAnInvalidSiteURL(t *testing.T) {
	t.Parallel()

	_, err := Load(env(map[string]string{"SITE_URL": "not a url"}))
	if err == nil {
		t.Fatal("Load() error = nil, want an error for an invalid SITE_URL")
	}
	if !strings.Contains(err.Error(), "SITE_URL") {
		t.Errorf("Load() error = %q, want it to name SITE_URL", err.Error())
	}
}
