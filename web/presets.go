package web

import (
	"fmt"
	"time"
)

const (
	RunModeAuto     = "auto"
	RunModeQuick    = "quick"
	RunModeBalanced = "balanced"
	RunModeDeep     = "deep"
	RunModeCustom   = "custom"
)

// EffectiveRunMode keeps pre-PinHarbor jobs truthful: historical jobs with a
// stored duration are treated as custom timed runs, while new empty jobs use
// Auto Mode.
func EffectiveRunMode(data JobData) string {
	if data.RunMode != "" {
		return data.RunMode
	}
	if data.MaxTime > 0 {
		return RunModeCustom
	}

	return RunModeAuto
}

// IsRunMode reports whether a preset name is supported by the web runner.
func IsRunMode(value string) bool {
	switch value {
	case RunModeAuto, RunModeQuick, RunModeBalanced, RunModeDeep, RunModeCustom:
		return true
	default:
		return false
	}
}

// ApplyRunPreset applies server-authoritative limits for the selected preset.
// Auto mode has no clock deadline and scrolls until Google Maps stops returning
// additional results; the scraper's inactivity guard still handles stalled work.
func ApplyRunPreset(data *JobData) error {
	data.RunMode = EffectiveRunMode(*data)
	if data.RunMode != RunModeCustom {
		data.FastMode = false
	}

	switch data.RunMode {
	case RunModeAuto:
		data.Depth = 0
		data.MaxTime = 0
		data.Email = true
		data.ExtraReviews = false
	case RunModeQuick:
		data.Depth = 25
		data.MaxTime = 15 * time.Minute
		data.Email = false
		data.ExtraReviews = false
	case RunModeBalanced:
		data.Depth = 100
		data.MaxTime = time.Hour
		data.Email = true
		data.ExtraReviews = false
	case RunModeDeep:
		data.Depth = 250
		data.MaxTime = 3 * time.Hour
		data.Email = true
		data.ExtraReviews = true
	case RunModeCustom:
		// Custom settings are supplied by the form or API unchanged.
	default:
		return fmt.Errorf("invalid run mode %q", data.RunMode)
	}

	return nil
}
