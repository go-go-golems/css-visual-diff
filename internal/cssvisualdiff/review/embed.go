//go:build embed

package review

import (
	"embed"
	"io/fs"
)

//go:embed embed/public
var embeddedFS embed.FS

// PublicFS is the SPA filesystem rooted at embed/public/.
// Vite builds to dist/ and cmd/build-web copies it to embed/public/.
var PublicFS, _ = fs.Sub(embeddedFS, "embed/public")
