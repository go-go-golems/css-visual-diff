package modes

import (
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

func prepareTarget(page *driver.Page, target config.Target) error {
	return service.PrepareTarget(page, target)
}

func runScriptPrepare(page *driver.Page, prepare *config.PrepareSpec) error {
	return service.RunScriptPrepare(page, prepare)
}

func runDirectReactGlobalPrepare(page *driver.Page, prepare *config.PrepareSpec) error {
	return service.RunDirectReactGlobalPrepare(page, prepare)
}

type directReactGlobalPrepareResult = service.DirectReactGlobalPrepareResult

func buildDirectReactGlobalScript(prepare *config.PrepareSpec) (string, error) {
	return service.BuildDirectReactGlobalScript(prepare)
}
