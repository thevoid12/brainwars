package assests

import "embed"

var (
	//go:embed css/*.css js/*.js  img/*.jpg img/*.svg ext/*.js
	AssestFS embed.FS
)
