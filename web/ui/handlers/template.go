package handlers

import (
	logs "brainwars/pkg/logger"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// RenderTemplate is a helper function to render templates with layout.html
func RenderTemplate(c *gin.Context, pageTemplate string, data gin.H) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	// Define the paths to the layout and page templates
	templatePaths := []string{
		filepath.Join(viper.GetString("app.uiTemplates"), "layout.html"),
		filepath.Join(viper.GetString("app.uiTemplates"), pageTemplate), // Load the actual page template
	}

	// Parse the templates
	tmpl, err := template.ParseFiles(templatePaths...)
	if err != nil {
		l.Sugar().Error("ParseFiles failed:", err)
		// RenderErrorTemplate(c, "Internal server error occurred", err)
		return
	}

	// Execute the layout template, which will pull the content block from the page template
	err = tmpl.ExecuteTemplate(c.Writer, "layout.html", data)
	if err != nil {
		l.Sugar().Error("ExecuteTemplate failed:", err)
		// RenderErrorTemplate(c, "Internal server error occurred", err)
		return
	}
}
