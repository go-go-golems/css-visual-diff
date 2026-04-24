package service

type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type StyleSnapshot struct {
	Exists     bool              `json:"exists"`
	Computed   map[string]string `json:"computed"`
	Bounds     *Bounds           `json:"bounds,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type ProbeSpec struct {
	Name       string   `json:"name"`
	Selector   string   `json:"selector"`
	Props      []string `json:"props,omitempty"`
	Attributes []string `json:"attributes,omitempty"`
	Source     string   `json:"source,omitempty"`
	Required   bool     `json:"required,omitempty"`
}

type SelectorStatus struct {
	Name      string  `json:"name,omitempty"`
	Selector  string  `json:"selector"`
	Source    string  `json:"source,omitempty"`
	Exists    bool    `json:"exists"`
	Visible   bool    `json:"visible"`
	Bounds    *Bounds `json:"bounds,omitempty"`
	TextStart string  `json:"text_start,omitempty"`
	Error     string  `json:"error,omitempty"`
}
