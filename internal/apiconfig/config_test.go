package apiconfig

import (
	"log/slog"
	"strings"
	"testing"
)

// env builds a getenv function over a fixed map. A key absent from the map
// behaves like an unset environment variable.
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
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
	if cfg.AllowPlainHTTP {
		t.Error("AllowPlainHTTP is on by default")
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v, want none", cfg.TrustedProxies)
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
			name: "ALLOW_PLAIN_HTTP",
			env:  map[string]string{"ALLOW_PLAIN_HTTP": "true"},
			check: func(t *testing.T, cfg Config) {
				if !cfg.AllowPlainHTTP {
					t.Error("AllowPlainHTTP = false, want true")
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
		{
			name: "TRUSTED_PROXIES",
			env:  map[string]string{"TRUSTED_PROXIES": "10.0.0.0/8, 192.0.2.1"},
			check: func(t *testing.T, cfg Config) {
				if len(cfg.TrustedProxies) != 2 {
					t.Errorf("TrustedProxies = %v, want two", cfg.TrustedProxies)
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

func TestLoadRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "log level", env: map[string]string{"LOG_LEVEL": "verbose"}, want: "verbose"},
		{name: "proxy list", env: map[string]string{"TRUSTED_PROXIES": "not-a-cidr"}, want: "trusted proxy"},
		{name: "plain http flag", env: map[string]string{"ALLOW_PLAIN_HTTP": "maybe"}, want: "ALLOW_PLAIN_HTTP"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(env(tc.env))
			if err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
			if !strings.HasPrefix(err.Error(), "apiconfig:") {
				t.Errorf("Load() error = %q, want it wrapped with \"apiconfig:\"", err.Error())
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Load() error = %q, want it to name %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	t.Parallel()

	_, err := Load(env(map[string]string{"LOG_LEVEL": "verbose", "TRUSTED_PROXIES": "not-a-cidr"}))
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
	for _, want := range []string{"verbose", "trusted proxy"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Load() error = %q, want it to name %q", err.Error(), want)
		}
	}
}
