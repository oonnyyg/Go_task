package env

import (
	"testing"
)

// TestCustomEnvironmentReplacesHostEnv ensures that a custom Options.Environment
// is the sole source of env values, replacing os.Environ(). A variable present
// in the host process environment but absent from the custom map must be
// treated as missing (and thus fall back to its envDefault when present).
func TestCustomEnvironmentReplacesHostEnv(t *testing.T) {
	type config struct {
		HostOnly string `env:"REPRO_HOST_ONLY"`
		WithDef  string `env:"REPRO_WITH_DEF,required" envDefault:"fallback"`
	}

	// Pollute the host process environment.
	t.Setenv("REPRO_HOST_ONLY", "from-host")

	// Custom environment deliberately omits REPRO_HOST_ONLY.
	var cfg config
	err := ParseWithOptions(&cfg, Options{
		Environment: map[string]string{
			"REPRO_WITH_DEF": "",
		},
	})
	isNoErr(t, err)

	// Host env must NOT leak into the parsed config.
	isEqual(t, "", cfg.HostOnly)
	// A missing key with a default must fall back to the default.
	isEqual(t, "fallback", cfg.WithDef)
}