package modes

import "testing"

func TestComputeSpecificity(t *testing.T) {
	cases := []struct {
		name     string
		selector string
		expected Specificity
	}{
		{
			name:     "id selector",
			selector: "#hero",
			expected: Specificity{A: 1, B: 0, C: 0},
		},
		{
			name:     "class selector",
			selector: ".button",
			expected: Specificity{A: 0, B: 1, C: 0},
		},
		{
			name:     "type selector",
			selector: "div",
			expected: Specificity{A: 0, B: 0, C: 1},
		},
		{
			name:     "compound selector",
			selector: "div#hero.button:hover",
			expected: Specificity{A: 1, B: 2, C: 1},
		},
		{
			name:     "combinators and pseudo-element",
			selector: "ul li.active > a::before",
			expected: Specificity{A: 0, B: 1, C: 4},
		},
		{
			name:     "attribute selectors",
			selector: "[data-foo][data-bar]",
			expected: Specificity{A: 0, B: 2, C: 0},
		},
		{
			name:     "not selector picks max",
			selector: ":not(.a, #b)",
			expected: Specificity{A: 1, B: 0, C: 0},
		},
	}

	for _, tc := range cases {
		got := computeSpecificity(tc.selector)
		if got != tc.expected {
			t.Fatalf("%s: expected %v got %v", tc.name, tc.expected, got)
		}
	}
}

func TestSelectWinner(t *testing.T) {
	cases := []struct {
		name       string
		candidates []Candidate
		expected   Winner
	}{
		{
			name: "important wins",
			candidates: []Candidate{
				{Selector: ".a", Value: "1", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}, Order: 1},
				{Selector: ".b", Value: "2", Important: true, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}, Order: 2},
			},
			expected: Winner{Selector: ".b", Value: "2", Important: true, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}},
		},
		{
			name: "origin beats specificity when importance ties",
			candidates: []Candidate{
				{Selector: ".author", Value: "1", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 10, C: 0}, Order: 2},
				{Selector: "style attribute", Value: "2", Important: false, Origin: OriginInline, Specificity: Specificity{A: 1, B: 0, C: 0}, Order: 1},
			},
			expected: Winner{Selector: "style attribute", Value: "2", Important: false, Origin: OriginInline, Specificity: Specificity{A: 1, B: 0, C: 0}},
		},
		{
			name: "specificity beats order",
			candidates: []Candidate{
				{Selector: ".a", Value: "1", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}, Order: 10},
				{Selector: "#id", Value: "2", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 1, B: 0, C: 0}, Order: 1},
			},
			expected: Winner{Selector: "#id", Value: "2", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 1, B: 0, C: 0}},
		},
		{
			name: "order breaks ties",
			candidates: []Candidate{
				{Selector: ".a", Value: "1", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}, Order: 1},
				{Selector: ".b", Value: "2", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}, Order: 2},
			},
			expected: Winner{Selector: ".b", Value: "2", Important: false, Origin: OriginAuthor, Specificity: Specificity{A: 0, B: 1, C: 0}},
		},
	}

	for _, tc := range cases {
		got := selectWinner(tc.candidates)
		if got != tc.expected {
			t.Fatalf("%s: expected %#v got %#v", tc.name, tc.expected, got)
		}
	}
}
