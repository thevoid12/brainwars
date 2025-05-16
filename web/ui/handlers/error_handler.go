package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Helper function to render the error template
func RenderErrorTemplate(c *gin.Context, pageTemplate, errorMessage string, errormsg error) {
	// Define the paths to the layout and page templates
	templatePaths := []string{
		filepath.Join(viper.GetString("app.uiTemplates"), "layout.html"),
		filepath.Join(viper.GetString("app.uiTemplates"), pageTemplate),
	}

	// Parse templates with error handling
	tmpl, err := template.ParseFiles(templatePaths...)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		c.String(http.StatusInternalServerError, "Template parsing failed")
		return
	}

	actualErr := "| Actual: "
	if errormsg != nil {
		actualErr += errormsg.Error()
	} else {
		actualErr = ""
	}

	data := map[string]interface{}{
		"Title":        "Error",
		"ErrorMessage": errorMessage + actualErr,
	}

	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		c.String(http.StatusInternalServerError, "Template execution failed")
		return
	}
}
