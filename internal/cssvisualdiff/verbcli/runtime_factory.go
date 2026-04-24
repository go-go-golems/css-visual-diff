package verbcli

import (
	"fmt"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl"
	"github.com/go-go-golems/go-go-goja/engine"
)

func newRuntimeFactory(repo ScannedRepository) (*engine.Factory, error) {
	if repo.Registry == nil {
		return nil, fmt.Errorf("repository %s has no jsverbs registry", describeRepository(repo))
	}
	return dsl.NewRuntimeFactory(repo.Registry, repo.RuntimeOptions()...)
}
