package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/adityayuga/signalfx-go-tracing/internal/globalconfig"
)

func TestAnalyticsSettings(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg := newConfig()
		assert.Equal(t, 0.0, cfg.analyticsRate)
	})

	t.Run("global", func(t *testing.T) {
		t.Skip("global flag disabled")
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := newConfig()
		assert.Equal(t, 0.4, cfg.analyticsRate)
	})

	t.Run("enabled", func(t *testing.T) {
		cfg := newConfig(WithAnalytics(true))
		assert.Equal(t, 1.0, cfg.analyticsRate)
	})

	t.Run("override", func(t *testing.T) {
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := newConfig(WithAnalyticsRate(0.2))
		assert.Equal(t, 0.2, cfg.analyticsRate)
	})
}
