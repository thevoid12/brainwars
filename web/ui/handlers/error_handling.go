package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Helper function to render the error template
func RenderErrorTemplate(c *gin.Context, errorMessage string, errormsg error) {
	layoutTmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "layout.html"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
		return
	}

	actualerr := ""
	if errormsg != nil {
		actualerr = errormsg.Error()
	}
	data := gin.H{
		"Title":        "Error",
		"ErrorMessage": errorMessage + "    Actual error: " + actualerr,
	}

	err = layoutTmpl.Execute(c.Writer, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute template"})
		return
	}
}
