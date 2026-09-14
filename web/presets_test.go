package web

import (
	"testing"
	"time"
)

func TestApplyRunPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mode        string
		wantDepth   int
		wantMaxTime time.Duration
		wantEmail   bool
		wantReviews bool
	}{
		{name: "auto", mode: RunModeAuto, wantDepth: 0, wantMaxTime: 0, wantEmail: true},
		{name: "quick", mode: RunModeQuick, wantDepth: 25, wantMaxTime: 15 * time.Minute},
		{name: "balanced", mode: RunModeBalanced, wantDepth: 100, wantMaxTime: time.Hour, wantEmail: true},
		{name: "deep", mode: RunModeDeep, wantDepth: 250, wantMaxTime: 3 * time.Hour, wantEmail: true, wantReviews: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			data := JobData{RunMode: test.mode, FastMode: true}
			if err := ApplyRunPreset(&data); err != nil {
				t.Fatalf("ApplyRunPreset: %v", err)
			}
			if data.Depth != test.wantDepth || data.MaxTime != test.wantMaxTime || data.Email != test.wantEmail || data.ExtraReviews != test.wantReviews {
				t.Fatalf("unexpected preset data: %+v", data)
			}
			if data.FastMode {
				t.Fatal("managed presets must disable Fast Mode")
			}
		})
	}
}

func TestCustomPresetPreservesManualValues(t *testing.T) {
	t.Parallel()

	data := JobData{
		RunMode:      RunModeCustom,
		Depth:        42,
		MaxTime:      20 * time.Minute,
		FastMode:     true,
		Email:        true,
		ExtraReviews: true,
	}
	if err := ApplyRunPreset(&data); err != nil {
		t.Fatalf("ApplyRunPreset: %v", err)
	}
	if data.Depth != 42 || data.MaxTime != 20*time.Minute || !data.FastMode || !data.Email || !data.ExtraReviews {
		t.Fatalf("custom values changed: %+v", data)
	}
}

func TestEffectiveRunModePreservesLegacyTimedJobs(t *testing.T) {
	t.Parallel()

	if got := EffectiveRunMode(JobData{MaxTime: time.Hour}); got != RunModeCustom {
		t.Fatalf("legacy timed job mode = %q, want custom", got)
	}
	if got := EffectiveRunMode(JobData{}); got != RunModeAuto {
		t.Fatalf("empty job mode = %q, want auto", got)
	}
}

func TestAutoModeValidationAllowsCompletionDrivenLimits(t *testing.T) {
	t.Parallel()

	data := JobData{
		Keywords: []string{"coffee"},
		RunMode:  RunModeAuto,
		Lang:     "en",
	}
	if err := data.Validate(); err != nil {
		t.Fatalf("auto mode should allow zero depth and no max time: %v", err)
	}
}
