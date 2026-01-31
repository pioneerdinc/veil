package quick

import (
	"strings"
	"time"
)

// ApplyTemplate replaces placeholders in a template string with values from Result
// Supported placeholders:
//   - {value} - the generated secret
//   - {type} - type of secret (password, apikey, jwt)
//   - {timestamp} - ISO timestamp
func ApplyTemplate(template string, result *Result) string {
	if template == "" {
		return result.Value
	}

	output := template
	output = strings.ReplaceAll(output, "{value}", result.Value)
	output = strings.ReplaceAll(output, "{type}", result.Type)
	output = strings.ReplaceAll(output, "{timestamp}", result.Timestamp.Format(time.RFC3339))

	return output
}

// FormatOutput returns the formatted output string for a result
// If a template is provided, it applies the template
// Otherwise, returns a default format like "Generated <type>: <value>"
func FormatOutput(result *Result, template string) string {
	if template != "" {
		return ApplyTemplate(template, result)
	}

	typeLabel := result.Type
	if typeLabel == "" {
		typeLabel = "password"
	}

	return "Generated " + typeLabel + ": " + result.Value
}
