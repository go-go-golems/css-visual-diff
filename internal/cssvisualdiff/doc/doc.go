package doc

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed topics/*.md examples/*.md tutorials/*.md
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
