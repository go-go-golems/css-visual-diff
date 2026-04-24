package dsl

import "embed"

//go:embed scripts/*.js
var embeddedScripts embed.FS
