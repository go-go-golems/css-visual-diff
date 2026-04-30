package service

// Viewport is the browser viewport used when loading a page.
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// PrepareSpec describes optional page preparation after navigation.
type PrepareSpec struct {
	Type             string         `json:"type"`
	Script           string         `json:"script"`
	ScriptFile       string         `json:"scriptFile"`
	WaitFor          string         `json:"waitFor"`
	WaitForTimeoutMS int            `json:"waitForTimeoutMs"`
	AfterWaitMS      int            `json:"afterWaitMs"`
	Component        string         `json:"component"`
	Props            map[string]any `json:"props"`
	RootSelector     string         `json:"rootSelector"`
	Width            int            `json:"width"`
	MinHeight        int            `json:"minHeight"`
	Background       string         `json:"background"`
}

// PageTarget is what the browser service needs to load and prepare a page.
type PageTarget struct {
	Name         string       `json:"name"`
	URL          string       `json:"url"`
	WaitMS       int          `json:"waitMs"`
	Viewport     Viewport     `json:"viewport"`
	RootSelector string       `json:"rootSelector,omitempty"`
	Prepare      *PrepareSpec `json:"prepare,omitempty"`
}
