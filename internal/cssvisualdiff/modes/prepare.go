package modes

import (
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

func prepareTarget(page *driver.Page, target config.Target) error {
	return service.PrepareTarget(page, toServicePageTarget(target))
}

//nolint:unused // Kept as a modes-level compatibility wrapper around the service package.
func runScriptPrepare(page *driver.Page, prepare *config.PrepareSpec) error {
	return service.RunScriptPrepare(page, toServicePrepareSpec(prepare))
}

//nolint:unused // Kept as a modes-level compatibility wrapper around the service package.
func runDirectReactGlobalPrepare(page *driver.Page, prepare *config.PrepareSpec) error {
	return service.RunDirectReactGlobalPrepare(page, toServicePrepareSpec(prepare))
}

//nolint:unused // Kept as a modes-level compatibility alias around the service package.
type directReactGlobalPrepareResult = service.DirectReactGlobalPrepareResult

func buildDirectReactGlobalScript(prepare *config.PrepareSpec) (string, error) {
	return service.BuildDirectReactGlobalScript(toServicePrepareSpec(prepare))
}
