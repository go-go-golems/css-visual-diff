package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/modes"
)

type AgentBriefOptions struct {
	Question   string
	Evidence   modes.CompareResult
	MaxBullets int
}

type AgentBriefResult struct {
	Question             string            `json:"question"`
	Bullets              []string          `json:"bullets"`
	PixelDiffPercent     float64           `json:"pixelDiffPercent"`
	ChangedPropertyCount int               `json:"changedPropertyCount"`
	Artifacts            map[string]string `json:"artifacts,omitempty"`
}

func BuildAgentBrief(opts AgentBriefOptions) AgentBriefResult {
	maxBullets := opts.MaxBullets
	if maxBullets <= 0 {
		maxBullets = 8
	}

	bullets := []string{}
	if opts.Evidence.PixelDiff.ChangedPercent > 0 {
		bullets = append(bullets, fmt.Sprintf("Visual drift is %.2f%% of pixels at threshold %d.", opts.Evidence.PixelDiff.ChangedPercent, opts.Evidence.PixelDiff.Threshold))
	}

	for _, diff := range opts.Evidence.ComputedDiffs {
		if strings.TrimSpace(diff.Original) == strings.TrimSpace(diff.React) {
			continue
		}
		bullets = append(bullets, fmt.Sprintf("Change `%s` from `%s` to `%s`.", diff.Property, safeValue(diff.Original), safeValue(diff.React)))
		if len(bullets) >= maxBullets {
			break
		}
	}

	if len(bullets) < maxBullets {
		winnerBullets := summarizeWinnerDiffs(opts.Evidence.WinnerDiffs)
		for _, bullet := range winnerBullets {
			bullets = append(bullets, bullet)
			if len(bullets) >= maxBullets {
				break
			}
		}
	}

	artifacts := map[string]string{}
	if opts.Evidence.URL1.ElementScreenshot != "" {
		artifacts["left"] = opts.Evidence.URL1.ElementScreenshot
	}
	if opts.Evidence.URL2.ElementScreenshot != "" {
		artifacts["right"] = opts.Evidence.URL2.ElementScreenshot
	}
	if opts.Evidence.PixelDiff.DiffComparisonPath != "" {
		artifacts["comparison"] = opts.Evidence.PixelDiff.DiffComparisonPath
	}
	if opts.Evidence.PixelDiff.DiffOnlyPath != "" {
		artifacts["diff"] = opts.Evidence.PixelDiff.DiffOnlyPath
	}
	if len(artifacts) == 0 {
		artifacts = nil
	}

	return AgentBriefResult{
		Question:             strings.TrimSpace(opts.Question),
		Bullets:              bullets,
		PixelDiffPercent:     opts.Evidence.PixelDiff.ChangedPercent,
		ChangedPropertyCount: len(opts.Evidence.ComputedDiffs),
		Artifacts:            artifacts,
	}
}

func RenderAgentBriefText(result AgentBriefResult) string {
	var lines []string
	if strings.TrimSpace(result.Question) != "" {
		lines = append(lines, result.Question)
	}
	for _, bullet := range result.Bullets {
		lines = append(lines, "- "+bullet)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func summarizeWinnerDiffs(diffs []modes.WinnerDiff) []string {
	bullets := []string{}
	seen := map[string]struct{}{}
	for _, diff := range diffs {
		left := strings.TrimSpace(diff.Original.Selector)
		right := strings.TrimSpace(diff.React.Selector)
		if left == right {
			continue
		}
		if left == "" && right == "" {
			continue
		}
		key := diff.Property + "|" + left + "|" + right
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		bullets = append(bullets, fmt.Sprintf("Winning rule for `%s` changed from `%s` to `%s`.", diff.Property, safeValue(left), safeValue(right)))
	}
	sort.Strings(bullets)
	return bullets
}

func safeValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	return value
}
