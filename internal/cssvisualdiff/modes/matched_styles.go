package modes

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/rs/zerolog/log"
)

type MatchedStylesResult struct {
	Styles []MatchedStyleEntry `json:"styles"`
}

type MatchedStyleEntry struct {
	Name             string          `json:"name"`
	Selector         string          `json:"selector,omitempty"`
	OriginalSelector string          `json:"original_selector,omitempty"`
	ReactSelector    string          `json:"react_selector,omitempty"`
	Original         MatchedSnapshot `json:"original"`
	React            MatchedSnapshot `json:"react"`
	Winners          []WinnerDiff    `json:"winners"`
}

type MatchedSnapshot struct {
	Exists   bool              `json:"exists"`
	Rules    []MatchedRule     `json:"rules"`
	Computed map[string]string `json:"computed"`
	Bounds   *Bounds           `json:"bounds,omitempty"`
}

type MatchedRule struct {
	Selector    string         `json:"selector"`
	Origin      CascadeOrigin  `json:"origin"`
	Specificity Specificity    `json:"specificity"`
	Properties  []RuleProperty `json:"properties"`
}

type RuleProperty struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Important bool   `json:"important"`
}

type Winner struct {
	Selector    string        `json:"selector"`
	Value       string        `json:"value"`
	Important   bool          `json:"important"`
	Origin      CascadeOrigin `json:"origin"`
	Specificity Specificity   `json:"specificity"`
}

type WinnerDiff struct {
	Property string `json:"property"`
	Original Winner `json:"original"`
	React    Winner `json:"react"`
}

type Specificity struct {
	A int `json:"a"`
	B int `json:"b"`
	C int `json:"c"`
}

func (s Specificity) Compare(other Specificity) int {
	if s.A != other.A {
		if s.A > other.A {
			return 1
		}
		return -1
	}
	if s.B != other.B {
		if s.B > other.B {
			return 1
		}
		return -1
	}
	if s.C != other.C {
		if s.C > other.C {
			return 1
		}
		return -1
	}
	return 0
}

func (s Specificity) String() string {
	return fmt.Sprintf("%d,%d,%d", s.A, s.B, s.C)
}

func (s Specificity) Add(other Specificity) Specificity {
	return Specificity{
		A: s.A + other.A,
		B: s.B + other.B,
		C: s.C + other.C,
	}
}

type CascadeOrigin string

const (
	OriginInline    CascadeOrigin = "inline"
	OriginAuthor    CascadeOrigin = "author"
	OriginUserAgent CascadeOrigin = "user-agent"
)

type Candidate struct {
	Property    string
	Value       string
	Selector    string
	Important   bool
	Specificity Specificity
	Origin      CascadeOrigin
	Order       int
}

func evaluateMatched(page *driver.Page, spec service.StyleEvalSpec) (MatchedSnapshot, error) {
	var nodeIDs []cdp.NodeID
	log.Info().Str("selector", spec.Selector).Msg("css-visual-diff matched-styles: query node IDs")
	if err := chromedp.Run(page.Context(), chromedp.NodeIDs(spec.Selector, &nodeIDs, chromedp.ByQuery)); err != nil {
		log.Error().Err(err).Str("selector", spec.Selector).Msg("css-visual-diff matched-styles: query node IDs failed")
		return MatchedSnapshot{}, err
	}
	if len(nodeIDs) == 0 {
		log.Info().Str("selector", spec.Selector).Msg("css-visual-diff matched-styles: selector not found")
		return MatchedSnapshot{Exists: false, Rules: []MatchedRule{}, Computed: map[string]string{}}, nil
	}
	nodeID := nodeIDs[0]

	log.Info().Int64("node_id", int64(nodeID)).Msg("css-visual-diff matched-styles: get matched styles")
	var inlineStyle *css.Style
	var matchedRules []*css.RuleMatch
	if err := chromedp.Run(page.Context(), chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		inlineStyle, _, matchedRules, _, _, _, _, _, _, _, _, _, _, _, err = css.GetMatchedStylesForNode(nodeID).Do(ctx)
		return err
	})); err != nil {
		log.Error().Err(err).Int64("node_id", int64(nodeID)).Msg("css-visual-diff matched-styles: get matched styles failed")
		return MatchedSnapshot{}, err
	}
	log.Info().Int64("node_id", int64(nodeID)).Msg("css-visual-diff matched-styles: get computed styles")
	var computedProps []*css.ComputedStyleProperty
	if err := chromedp.Run(page.Context(), chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		computedProps, err = css.GetComputedStyleForNode(nodeID).Do(ctx)
		return err
	})); err != nil {
		log.Error().Err(err).Int64("node_id", int64(nodeID)).Msg("css-visual-diff matched-styles: get computed styles failed")
		return MatchedSnapshot{}, err
	}

	computedMap := map[string]string{}
	for _, prop := range computedProps {
		computedMap[prop.Name] = prop.Value
	}

	var bounds *Bounds
	if hasReport(spec.Report, "box_model") {
		box, err := dom.GetBoxModel().WithNodeID(nodeID).Do(page.Context())
		if err == nil && box != nil {
			bounds = boxModelBounds(box)
		}
	}

	rules := []MatchedRule{}
	for _, ruleMatch := range matchedRules {
		selector, specificity := selectorFromRuleMatch(ruleMatch)
		props := []RuleProperty{}
		if ruleMatch.Rule != nil && ruleMatch.Rule.Style != nil {
			for _, prop := range ruleMatch.Rule.Style.CSSProperties {
				props = append(props, RuleProperty{
					Name:      prop.Name,
					Value:     prop.Value,
					Important: prop.Important,
				})
			}
		}
		rules = append(rules, MatchedRule{
			Selector:    selector,
			Origin:      originFromRule(ruleMatch),
			Specificity: specificity,
			Properties:  props,
		})
	}

	if inlineStyle != nil && len(inlineStyle.CSSProperties) > 0 {
		inlineProps := []RuleProperty{}
		for _, prop := range inlineStyle.CSSProperties {
			inlineProps = append(inlineProps, RuleProperty{
				Name:      prop.Name,
				Value:     prop.Value,
				Important: prop.Important,
			})
		}
		rules = append(rules, MatchedRule{
			Selector:    "style attribute",
			Origin:      OriginInline,
			Specificity: Specificity{A: 1, B: 0, C: 0},
			Properties:  inlineProps,
		})
	}

	return MatchedSnapshot{Exists: true, Rules: rules, Computed: computedMap, Bounds: bounds}, nil
}

func buildWinnerDiffs(props []string, orig MatchedSnapshot, react MatchedSnapshot) []WinnerDiff {
	var diffs []WinnerDiff
	origCandidates := collectCandidates(orig.Rules)
	reactCandidates := collectCandidates(react.Rules)
	for _, prop := range props {
		origWinner := selectWinner(origCandidates[prop])
		reactWinner := selectWinner(reactCandidates[prop])
		diffs = append(diffs, WinnerDiff{Property: prop, Original: origWinner, React: reactWinner})
	}
	return diffs
}

func collectCandidates(rules []MatchedRule) map[string][]Candidate {
	candidates := map[string][]Candidate{}
	order := 0
	for _, rule := range rules {
		for _, prop := range rule.Properties {
			order++
			candidate := Candidate{
				Property:    prop.Name,
				Value:       prop.Value,
				Selector:    rule.Selector,
				Important:   prop.Important,
				Specificity: rule.Specificity,
				Origin:      rule.Origin,
				Order:       order,
			}
			candidates[prop.Name] = append(candidates[prop.Name], candidate)
		}
	}
	return candidates
}

func selectWinner(candidates []Candidate) Winner {
	if len(candidates) == 0 {
		return Winner{}
	}
	winner := candidates[0]
	for i := 1; i < len(candidates); i++ {
		if candidateBeats(candidates[i], winner) {
			winner = candidates[i]
		}
	}
	return Winner{
		Selector:    winner.Selector,
		Value:       winner.Value,
		Important:   winner.Important,
		Origin:      winner.Origin,
		Specificity: winner.Specificity,
	}
}

func candidateBeats(a Candidate, b Candidate) bool {
	if a.Important != b.Important {
		return a.Important && !b.Important
	}
	if originPriority(a.Origin) != originPriority(b.Origin) {
		return originPriority(a.Origin) > originPriority(b.Origin)
	}
	if cmp := a.Specificity.Compare(b.Specificity); cmp != 0 {
		return cmp > 0
	}
	return a.Order > b.Order
}

func originPriority(origin CascadeOrigin) int {
	switch origin {
	case OriginInline:
		return 3
	case OriginAuthor:
		return 2
	case OriginUserAgent:
		return 1
	default:
		return 2
	}
}

func boxModelBounds(model *dom.BoxModel) *Bounds {
	if model == nil || len(model.Content) < 8 {
		return nil
	}
	content := model.Content
	return &Bounds{
		X:      content[0],
		Y:      content[1],
		Width:  content[2] - content[0],
		Height: content[5] - content[1],
	}
}

func hasReport(reports []string, value string) bool {
	for _, r := range reports {
		if r == value {
			return true
		}
	}
	return false
}

func selectorFromRuleMatch(ruleMatch *css.RuleMatch) (string, Specificity) {
	if ruleMatch == nil || ruleMatch.Rule == nil || ruleMatch.Rule.SelectorList == nil {
		return "", Specificity{}
	}
	selectors := ruleMatch.Rule.SelectorList.Selectors
	if len(ruleMatch.MatchingSelectors) > 0 && len(selectors) > 0 {
		var bestSelector *css.Value
		bestSpecificity := Specificity{}
		hasBest := false
		for _, idx := range ruleMatch.MatchingSelectors {
			if idx < 0 || int(idx) >= len(selectors) {
				continue
			}
			value := selectors[idx]
			spec := specificityFromValue(value)
			if !hasBest || spec.Compare(bestSpecificity) > 0 {
				bestSelector = value
				bestSpecificity = spec
				hasBest = true
			}
		}
		if bestSelector != nil {
			return bestSelector.Text, bestSpecificity
		}
	}
	if len(selectors) > 0 {
		return selectors[0].Text, specificityFromValue(selectors[0])
	}
	text := ruleMatch.Rule.SelectorList.Text
	return text, specificityForSelectorList(text)
}

func specificityFromValue(value *css.Value) Specificity {
	if value == nil {
		return Specificity{}
	}
	if value.Specificity != nil {
		return Specificity{
			A: int(value.Specificity.A),
			B: int(value.Specificity.B),
			C: int(value.Specificity.C),
		}
	}
	return computeSpecificity(value.Text)
}

func originFromRule(ruleMatch *css.RuleMatch) CascadeOrigin {
	if ruleMatch == nil || ruleMatch.Rule == nil {
		return OriginAuthor
	}
	switch ruleMatch.Rule.Origin {
	case css.StyleSheetOriginUserAgent:
		return OriginUserAgent
	case css.StyleSheetOriginRegular, css.StyleSheetOriginInjected, css.StyleSheetOriginInspector:
		return OriginAuthor
	default:
		return OriginAuthor
	}
}

func formatWinner(winner Winner) string {
	if strings.TrimSpace(winner.Selector) == "" && strings.TrimSpace(winner.Value) == "" {
		return "none"
	}
	important := ""
	if winner.Important {
		important = " !important"
	}
	return fmt.Sprintf("%s (%s%s; origin=%s; spec=%s)", winner.Selector, winner.Value, important, winner.Origin, winner.Specificity.String())
}

func specificityForSelectorList(selectorList string) Specificity {
	selectors := splitSelectorList(selectorList)
	best := Specificity{}
	hasBest := false
	for _, sel := range selectors {
		trimmed := strings.TrimSpace(sel)
		if trimmed == "" {
			continue
		}
		spec := computeSpecificity(trimmed)
		if !hasBest || spec.Compare(best) > 0 {
			best = spec
			hasBest = true
		}
	}
	if !hasBest {
		return Specificity{}
	}
	return best
}

func splitSelectorList(selectorList string) []string {
	var parts []string
	start := 0
	depthParen := 0
	depthBracket := 0
	inString := byte(0)
	for i := 0; i < len(selectorList); i++ {
		ch := selectorList[i]
		if inString != 0 {
			if ch == '\\' {
				i++
				continue
			}
			if ch == inString {
				inString = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			inString = ch
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case ',':
			if depthParen == 0 && depthBracket == 0 {
				parts = append(parts, selectorList[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, selectorList[start:])
	return parts
}

func computeSpecificity(selector string) Specificity {
	spec := Specificity{}
	expectType := true
	for i := 0; i < len(selector); {
		ch := selector[i]
		if isWhitespace(ch) {
			expectType = true
			i++
			continue
		}
		switch ch {
		case ',':
			expectType = true
			i++
		case '>', '+', '~':
			expectType = true
			i++
		case '#':
			spec.A++
			expectType = false
			i = skipIdent(selector, i+1)
		case '.':
			spec.B++
			expectType = false
			i = skipIdent(selector, i+1)
		case '[':
			spec.B++
			expectType = false
			_, next := scanEnclosed(selector, i, '[', ']')
			i = next
		case ':':
			expectType = false
			if i+1 < len(selector) && selector[i+1] == ':' {
				spec.C++
				i = skipIdent(selector, i+2)
				continue
			}
			name, next := readIdent(selector, i+1)
			lower := strings.ToLower(name)
			if isLegacyPseudoElement(lower) {
				spec.C++
				i = next
				continue
			}
			if lower == "where" {
				if next < len(selector) && selector[next] == '(' {
					_, end := scanEnclosed(selector, next, '(', ')')
					i = end
					continue
				}
				i = next
				continue
			}
			if lower == "not" || lower == "is" || lower == "has" {
				if next < len(selector) && selector[next] == '(' {
					content, end := scanEnclosed(selector, next, '(', ')')
					spec = spec.Add(specificityForSelectorList(content))
					i = end
					continue
				}
				spec.B++
				i = next
				continue
			}
			spec.B++
			if next < len(selector) && selector[next] == '(' {
				_, end := scanEnclosed(selector, next, '(', ')')
				i = end
				continue
			}
			i = next
		case '*':
			expectType = false
			i++
		default:
			if isIdentStart(ch) {
				if expectType {
					spec.C++
					expectType = false
				}
				i = skipIdent(selector, i+1)
				continue
			}
			i++
		}
	}
	return spec
}

func readIdent(input string, start int) (string, int) {
	i := start
	for i < len(input) && isIdentChar(input[i]) {
		i++
	}
	return input[start:i], i
}

func skipIdent(input string, start int) int {
	_, next := readIdent(input, start)
	return next
}

func scanEnclosed(input string, start int, open, closeByte byte) (string, int) {
	if start >= len(input) || input[start] != open {
		return "", start
	}
	depth := 1
	inString := byte(0)
	for i := start + 1; i < len(input); i++ {
		ch := input[i]
		if inString != 0 {
			if ch == '\\' {
				i++
				continue
			}
			if ch == inString {
				inString = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			inString = ch
		case open:
			depth++
		case closeByte:
			depth--
			if depth == 0 {
				return input[start+1 : i], i + 1
			}
		}
	}
	return input[start+1:], len(input)
}

func isLegacyPseudoElement(name string) bool {
	switch name {
	case "before", "after", "first-line", "first-letter":
		return true
	default:
		return false
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}

func isIdentStart(ch byte) bool {
	return ch == '_' || ch == '-' || unicode.IsLetter(rune(ch))
}

func isIdentChar(ch byte) bool {
	return ch == '_' || ch == '-' || unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch))
}
