package modes

import (
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

func prepareTarget(page *driver.Page, target config.Target) error {
	return service.PrepareTarget(page, target)
}

func buildDirectReactGlobalScript(prepare *config.PrepareSpec) (string, error) {
	return service.BuildDirectReactGlobalScript(prepare)
}
