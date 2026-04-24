package service

import (
	"context"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

type BrowserService struct {
	browser *driver.Browser
}

type PageService struct {
	page *driver.Page
}

func NewBrowserService(ctx context.Context) (*BrowserService, error) {
	browser, err := driver.NewBrowser(ctx)
	if err != nil {
		return nil, err
	}
	return &BrowserService{browser: browser}, nil
}

func (s *BrowserService) Browser() *driver.Browser {
	if s == nil {
		return nil
	}
	return s.browser
}

func (s *BrowserService) Close() {
	if s != nil && s.browser != nil {
		s.browser.Close()
	}
}

func (s *BrowserService) NewPage() (*PageService, error) {
	page, err := s.browser.NewPage()
	if err != nil {
		return nil, err
	}
	return &PageService{page: page}, nil
}

func (p *PageService) Page() *driver.Page {
	if p == nil {
		return nil
	}
	return p.page
}

func (p *PageService) Close() {
	if p != nil && p.page != nil {
		p.page.Close()
	}
}

func (p *PageService) LoadAndPrepareTarget(target config.Target) error {
	return LoadAndPreparePage(p.page, target)
}

func LoadAndPreparePage(page *driver.Page, target config.Target) error {
	if err := page.SetViewport(target.Viewport.Width, target.Viewport.Height); err != nil {
		return err
	}
	if err := page.Goto(target.URL); err != nil {
		return err
	}
	if target.WaitMS > 0 {
		page.Wait(time.Duration(target.WaitMS) * time.Millisecond)
	}
	return PrepareTarget(page, target)
}
