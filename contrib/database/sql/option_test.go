package sql

import (
	"testing"

	"github.com/adityayuga/signalfx-go-tracing/internal/globalconfig"
	"gotest.tools/assert"
)

func TestAnalyticsSettings(t *testing.T) {
	t.Run("global", func(t *testing.T) {
		t.Skip("global flag disabled")
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := new(registerConfig)
		defaults(cfg)
		assert.Equal(t, 0.4, cfg.analyticsRate)
	})

	t.Run("enabled", func(t *testing.T) {
		cfg := new(registerConfig)
		defaults(cfg)
		WithAnalytics(true)(cfg)
		assert.Equal(t, 1.0, cfg.analyticsRate)
	})

	t.Run("override", func(t *testing.T) {
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := new(registerConfig)
		defaults(cfg)
		WithAnalyticsRate(0.2)(cfg)
		assert.Equal(t, 0.2, cfg.analyticsRate)
	})
}
