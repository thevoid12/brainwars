package assests

import "embed"

var (
	//go:embed css/*.css js/*.js  img/*.jpg img/*.svg ext/*.js favicon_io/*ico
	AssestFS embed.FS
)
