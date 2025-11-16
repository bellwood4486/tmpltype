package main

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// GetTemplateFuncs returns custom template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// String manipulation
		"upper": strings.ToUpper,
		"lower": strings.ToLower,

		// Date formatting
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},

		// HTML formatting
		"nl2br": func(s string) template.HTML {
			escaped := template.HTMLEscapeString(s)
			return template.HTML(strings.ReplaceAll(escaped, "\n", "<br>"))
		},

		// Default value
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},

		// Number formatting
		"comma": func(n int) string {
			s := fmt.Sprintf("%d", n)
			if n < 1000 {
				return s
			}
			// Simple thousands separator
			result := ""
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += ","
				}
				result += string(c)
			}
			return result
		},
	}
}
