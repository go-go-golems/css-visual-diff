package service

import "github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"

type SnapshotProbeSpec struct {
	Name       string          `json:"name"`
	Selector   string          `json:"selector"`
	Source     string          `json:"source,omitempty"`
	Required   bool            `json:"required,omitempty"`
	Extractors []ExtractorSpec `json:"extractors"`
}

type ProbeSnapshot struct {
	Name     string          `json:"name"`
	Selector string          `json:"selector"`
	Source   string          `json:"source,omitempty"`
	Snapshot ElementSnapshot `json:"snapshot"`
	Error    string          `json:"error,omitempty"`
}

type PageSnapshot struct {
	Results []ProbeSnapshot `json:"results"`
}

func SnapshotPage(page *driver.Page, probes []SnapshotProbeSpec) (PageSnapshot, error) {
	results := make([]ProbeSnapshot, 0, len(probes))
	for _, probe := range probes {
		snapshot, err := ExtractElement(page, LocatorSpec{Name: probe.Name, Selector: probe.Selector, Source: probe.Source}, probe.Extractors)
		result := ProbeSnapshot{Name: probe.Name, Selector: probe.Selector, Source: probe.Source, Snapshot: snapshot}
		if err != nil {
			if probe.Required {
				return PageSnapshot{}, err
			}
			result.Error = err.Error()
		}
		results = append(results, result)
	}
	return PageSnapshot{Results: results}, nil
}
