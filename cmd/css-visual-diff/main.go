package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/llm"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/modes"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/runner"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type RunCommand struct {
	*cmds.CommandDescription
}

type RunSettings struct {
	Config             string `glazed.parameter:"config"`
	Modes              string `glazed.parameter:"modes"`
	DryRun             bool   `glazed.parameter:"dry-run"`
	PixelDiffThreshold int    `glazed.parameter:"pixeldiff-threshold"`
}

func NewRunCommand() (*RunCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"run",
		cmds.WithShort("Run css-visual-diff capture and analysis modes"),
		cmds.WithLong("Run css-visual-diff using a YAML capture plan and one or more modes."),
		cmds.WithFlags(
			fields.New(
				"config",
				fields.TypeString,
				fields.WithHelp("Path to css-visual-diff YAML config"),
			),
			fields.New(
				"modes",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Comma-delimited list of modes (capture,cssdiff,matched-styles,ai-review,full)"),
			),
			fields.New(
				"dry-run",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Validate config and modes without running browser actions"),
			),
			fields.New(
				"pixeldiff-threshold",
				fields.TypeInteger,
				fields.WithDefault(30),
				fields.WithHelp("Pixel diff threshold (0-255) used by pixeldiff mode"),
			),
		),
	)

	return &RunCommand{CommandDescription: cmdDesc}, nil
}

func (c *RunCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &RunSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.Config == "" {
		return fmt.Errorf("--config is required")
	}

	cfg, err := config.Load(settings.Config)
	if err != nil {
		return err
	}

	modesRaw := settings.Modes
	if modesRaw == "" && len(cfg.Modes) > 0 {
		modesRaw = joinModes(cfg.Modes)
	}

	modesList, err := runner.NormalizeModes(modesRaw)
	if err != nil {
		return err
	}

	result, err := runner.Run(ctx, cfg, modesList, settings.DryRun, runner.RunOptions{
		PixelDiffThreshold: settings.PixelDiffThreshold,
	})
	if err != nil && settings.DryRun {
		return err
	}

	for _, r := range result.Results {
		row := types.NewRow(
			types.MRP("mode", r.Mode),
			types.MRP("status", r.Status),
			types.MRP("message", r.Message),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	if !settings.DryRun {
		if containsMode(modesList, "capture") {
			if err := emitCoverageRows(ctx, gp, cfg.Output.Dir); err != nil {
				return err
			}
		}
		if containsMode(modesList, "story-discovery") {
			if err := emitStoryRows(ctx, gp, cfg.Output.Dir); err != nil {
				return err
			}
		}
	}

	return err
}

func joinModes(modes []string) string {
	out := ""
	for i, m := range modes {
		if i == 0 {
			out = m
			continue
		}
		out += "," + m
	}
	return out
}

func containsMode(modesList []string, value string) bool {
	for _, m := range modesList {
		if m == value {
			return true
		}
	}
	return false
}

func emitCoverageRows(ctx context.Context, gp middlewares.Processor, outDir string) error {
	data, err := os.ReadFile(filepath.Join(outDir, "capture.json"))
	if err != nil {
		return err
	}
	var capture modes.CaptureResult
	if err := json.Unmarshal(data, &capture); err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("type", "coverage"),
		types.MRP("total", capture.Coverage.Total),
		types.MRP("original_missing", capture.Coverage.OriginalMissing),
		types.MRP("react_missing", capture.Coverage.ReactMissing),
		types.MRP("original_hidden", capture.Coverage.OriginalHidden),
		types.MRP("react_hidden", capture.Coverage.ReactHidden),
	)
	return gp.AddRow(ctx, row)
}

func emitStoryRows(ctx context.Context, gp middlewares.Processor, outDir string) error {
	data, err := os.ReadFile(filepath.Join(outDir, "stories.json"))
	if err != nil {
		return err
	}
	var stories modes.StoryDiscoveryResult
	if err := json.Unmarshal(data, &stories); err != nil {
		return err
	}
	for _, entry := range stories.Entries {
		row := types.NewRow(
			types.MRP("type", "story"),
			types.MRP("id", entry.ID),
			types.MRP("title", entry.Title),
			types.MRP("name", entry.Name),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	runCmd, err := NewRunCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating run command: %v\n", err)
		os.Exit(1)
	}

	host, err := dsl.NewHost()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating script host: %v\n", err)
		os.Exit(1)
	}

	scriptCommands, err := host.Commands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building script commands: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "css-visual-diff",
		Short: "Compare rendered HTML/CSS across browser targets",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "css-visual-diff"); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding logging flags: %v\n", err)
		os.Exit(1)
	}
	setDefaultFlagValue(rootCmd, "log-level", "error")
	setDefaultFlagValue(rootCmd, "log-format", "text")

	cobraRunCmd, err := cli.BuildCobraCommand(runCmd, cli.WithParserConfig(cli.CobraParserConfig{
		ShortHelpSections: []string{schema.DefaultSlug},
		MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
	}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building run command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraRunCmd)
	rootCmd.AddCommand(newCompareCommand())
	rootCmd.AddCommand(newLLMReviewCommand())
	rootCmd.AddCommand(newChromedpProbeCommand())
	if err := cli.AddCommandsToRootCommand(rootCmd, scriptCommands, nil, cli.WithParserConfig(cli.CobraParserConfig{
		ShortHelpSections: []string{schema.DefaultSlug},
		MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
	})); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering script commands: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type compareSettings struct {
	URL1      string
	Selector1 string
	WaitMS1   int

	URL2      string
	Selector2 string
	WaitMS2   int

	ViewportW int
	ViewportH int

	Props string
	Attrs string

	OutDir string

	WriteJSON     bool
	WriteMarkdown bool
	WritePNGs     bool

	Threshold int
}

func defaultCompareProps() []string {
	return []string{
		"display",
		"position",
		"top",
		"right",
		"bottom",
		"left",
		"width",
		"height",
		"margin-top",
		"margin-right",
		"margin-bottom",
		"margin-left",
		"padding-top",
		"padding-right",
		"padding-bottom",
		"padding-left",
		"font-family",
		"font-size",
		"font-weight",
		"line-height",
		"color",
		"background-color",
		"background-image",
		"border-radius",
		"box-shadow",
		"z-index",
	}
}

func defaultCompareAttrs() []string {
	return []string{"id", "class"}
}

func buildCompareModeSettings(settings *compareSettings) modes.CompareSettings {
	props := parseCSV(settings.Props)
	if len(props) == 0 {
		props = defaultCompareProps()
	}

	attrs := parseCSV(settings.Attrs)
	if len(attrs) == 0 {
		attrs = defaultCompareAttrs()
	}

	return modes.CompareSettings{
		URL1:               settings.URL1,
		Selector1:          settings.Selector1,
		WaitMS1:            settings.WaitMS1,
		URL2:               settings.URL2,
		Selector2:          settings.Selector2,
		WaitMS2:            settings.WaitMS2,
		ViewportW:          settings.ViewportW,
		ViewportH:          settings.ViewportH,
		Props:              props,
		Attributes:         attrs,
		OutDir:             settings.OutDir,
		WriteJSON:          settings.WriteJSON,
		WriteMarkdown:      settings.WriteMarkdown,
		WritePNGs:          settings.WritePNGs,
		PixelDiffThreshold: settings.Threshold,
	}
}

func addCompareFlags(cmd *cobra.Command, settings *compareSettings) {
	cmd.Flags().StringVar(&settings.URL1, "url1", "", "First URL to compare")
	cmd.Flags().StringVar(&settings.Selector1, "selector1", "", "CSS selector for URL1 element")
	cmd.Flags().IntVar(&settings.WaitMS1, "wait-ms1", 0, "Wait after navigation for URL1 (ms)")

	cmd.Flags().StringVar(&settings.URL2, "url2", "", "Second URL to compare")
	cmd.Flags().StringVar(&settings.Selector2, "selector2", "", "CSS selector for URL2 element (defaults to selector1)")
	cmd.Flags().IntVar(&settings.WaitMS2, "wait-ms2", 0, "Wait after navigation for URL2 (ms)")

	cmd.Flags().IntVar(&settings.ViewportW, "viewport-w", 1280, "Viewport width")
	cmd.Flags().IntVar(&settings.ViewportH, "viewport-h", 720, "Viewport height")

	cmd.Flags().StringVar(&settings.Props, "props", "", "Comma-delimited CSS properties to compare (default: a curated list)")
	cmd.Flags().StringVar(&settings.Attrs, "attrs", "id,class", "Comma-delimited attributes to capture (default: id,class)")

	cmd.Flags().StringVar(&settings.OutDir, "out", "", "Output directory (default: ./css-visual-diff-compare-YYYYMMDD_HHMMSS)")
	cmd.Flags().IntVar(&settings.Threshold, "threshold", 30, "Pixel diff threshold (0-255)")

	cmd.Flags().BoolVar(&settings.WriteJSON, "write-json", true, "Write compare.json")
	cmd.Flags().BoolVar(&settings.WriteMarkdown, "write-markdown", true, "Write compare.md")
	cmd.Flags().BoolVar(&settings.WritePNGs, "write-pngs", true, "Write screenshots and diff images")

	_ = cmd.MarkFlagRequired("url1")
	_ = cmd.MarkFlagRequired("selector1")
	_ = cmd.MarkFlagRequired("url2")
}

func newCompareCommand() *cobra.Command {
	settings := &compareSettings{}
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare a single element between two URLs (screenshots + CSS + pixel diff)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return modes.Compare(cmd.Context(), buildCompareModeSettings(settings))
		},
	}

	addCompareFlags(cmd, settings)
	return cmd
}

type llmReviewSettings struct {
	compareSettings

	Question          string
	ConfigFile        string
	Profile           string
	ProfileRegistries []string

	WriteReviewJSON     bool
	WriteReviewMarkdown bool
}

func newLLMReviewCommand() *cobra.Command {
	settings := &llmReviewSettings{}
	cmd := &cobra.Command{
		Use:   "llm-review",
		Short: "Compare a region and ask an LLM for a multimodal review using Pinocchio profile settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			compareSettings := buildCompareModeSettings(&settings.compareSettings)
			result, err := modes.GenerateCompareResult(cmd.Context(), compareSettings)
			if err != nil {
				return err
			}
			if err := modes.WriteCompareArtifacts(result, compareSettings.WriteJSON, compareSettings.WriteMarkdown); err != nil {
				return err
			}

			bootstrapResult, err := llm.ResolveEngineSettings(cmd.Context(), llm.BootstrapOptions{
				ConfigFile:        settings.ConfigFile,
				Profile:           settings.Profile,
				ProfileRegistries: settings.ProfileRegistries,
			})
			if err != nil {
				return err
			}
			defer bootstrapResult.Close()

			review, err := llm.ReviewCompare(cmd.Context(), bootstrapResult, llm.ReviewOptions{
				Question: settings.Question,
				Evidence: result,
			})
			if err != nil {
				return err
			}

			if settings.WriteReviewJSON {
				if err := llm.WriteReviewJSON(filepath.Join(result.Inputs.OutDir, "llm-review.json"), review); err != nil {
					return err
				}
			}
			if settings.WriteReviewMarkdown {
				if err := llm.WriteReviewMarkdown(filepath.Join(result.Inputs.OutDir, "llm-review.md"), review); err != nil {
					return err
				}
			}

			fmt.Print(review.Answer)
			if !strings.HasSuffix(review.Answer, "\n") {
				fmt.Println()
			}
			return nil
		},
	}

	addCompareFlags(cmd, &settings.compareSettings)
	cmd.Flags().StringVar(&settings.Question, "question", "What are the main visual differences and their likely CSS causes?", "Question to ask about the compared regions")
	cmd.Flags().StringVar(&settings.ConfigFile, "config-file", "", "Optional Pinocchio config file used for profile/bootstrap resolution")
	cmd.Flags().StringVar(&settings.Profile, "profile", "", "Pinocchio/Geppetto engine profile to resolve")
	cmd.Flags().StringSliceVar(&settings.ProfileRegistries, "profile-registries", nil, "Comma-separated or repeated Pinocchio/Geppetto profile registry sources")
	cmd.Flags().BoolVar(&settings.WriteReviewJSON, "write-review-json", true, "Write llm-review.json")
	cmd.Flags().BoolVar(&settings.WriteReviewMarkdown, "write-review-markdown", true, "Write llm-review.md")
	return cmd
}

func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item == "" {
			continue
		}
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

type chromedpProbeSettings struct {
	URL       string
	Selector  string
	WaitMS    int
	ViewportW int
	ViewportH int
	TimeoutMS int
}

func setDefaultFlagValue(root *cobra.Command, name string, value string) {
	flag := root.PersistentFlags().Lookup(name)
	if flag == nil {
		return
	}
	flag.DefValue = value
	_ = flag.Value.Set(value)
}

func newChromedpProbeCommand() *cobra.Command {
	settings := &chromedpProbeSettings{}
	cmd := &cobra.Command{
		Use:   "chromedp-probe",
		Short: "Run a minimal chromedp probe against a URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if settings.URL == "" {
				return fmt.Errorf("--url is required")
			}
			timeout := time.Duration(settings.TimeoutMS) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			browser, err := driver.NewBrowser(ctx)
			if err != nil {
				return err
			}
			defer browser.Close()

			page, err := browser.NewPage()
			if err != nil {
				return err
			}
			defer page.Close()

			if err := page.SetViewport(settings.ViewportW, settings.ViewportH); err != nil {
				return err
			}
			if err := page.Goto(settings.URL); err != nil {
				return err
			}
			if settings.WaitMS > 0 {
				page.Wait(time.Duration(settings.WaitMS) * time.Millisecond)
			}

			var title string
			if err := page.Evaluate("document.title", &title); err != nil {
				return err
			}

			selectorMatches := -1
			if settings.Selector != "" {
				var nodeIDs []cdp.NodeID
				if err := chromedp.Run(page.Context(), chromedp.NodeIDs(settings.Selector, &nodeIDs, chromedp.ByQuery)); err != nil {
					return err
				}
				selectorMatches = len(nodeIDs)
			}

			if settings.Selector != "" {
				fmt.Printf("chromedp ok url=%s title=%q selector=%s matches=%d\n", settings.URL, title, settings.Selector, selectorMatches)
				return nil
			}
			fmt.Printf("chromedp ok url=%s title=%q\n", settings.URL, title)
			return nil
		},
	}

	cmd.Flags().StringVar(&settings.URL, "url", "", "URL to navigate to")
	cmd.Flags().StringVar(&settings.Selector, "selector", "", "Optional CSS selector to verify")
	cmd.Flags().IntVar(&settings.WaitMS, "wait-ms", 0, "Wait time in milliseconds after navigation")
	cmd.Flags().IntVar(&settings.ViewportW, "viewport-width", 1280, "Viewport width")
	cmd.Flags().IntVar(&settings.ViewportH, "viewport-height", 720, "Viewport height")
	cmd.Flags().IntVar(&settings.TimeoutMS, "timeout-ms", 30000, "Overall timeout in milliseconds")

	return cmd
}
