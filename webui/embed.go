// Package webui holds the cockpit's HTML/CSS/JS and exposes them as an
// embedded filesystem for release builds. The //go:embed directive has to live
// in a package at or above the assets, which is why this Go file sits in webui.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed index.html styles scripts
var assets embed.FS

// FS returns the embedded cockpit assets, rooted so that index.html is at the
// top level.
func FS() fs.FS { return assets }
