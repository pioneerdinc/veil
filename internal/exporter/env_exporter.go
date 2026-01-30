package exporter

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

type EnvExporter struct{}

type envLine struct {
	key       string
	value     string
	original  string
	isComment bool
	isEmpty   bool
}

func (e *EnvExporter) Format() string {
	return "env"
}

func (e *EnvExporter) Export(secrets map[string]string, opts ExportOptions) error {
	if !opts.Append && !opts.Force && fileExists(opts.TargetPath) {
		return fmt.Errorf("file %s already exists (use --force to overwrite or --append to add to it)", opts.TargetPath)
	}

	preview, err := e.Preview(secrets, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return nil
	}

	if opts.Append && fileExists(opts.TargetPath) {
		return e.appendToFile(secrets, preview, opts)
	}

	return e.writeNewFile(preview, opts)
}

func (e *EnvExporter) Preview(secrets map[string]string, opts ExportOptions) (*Preview, error) {
	preview := &Preview{
		NewKeys:     make([]string, 0),
		UpdatedKeys: make([]string, 0),
		SkippedKeys: make([]string, 0),
	}

	existingKeys := make(map[string]string)
	if opts.Append && fileExists(opts.TargetPath) {
		existingKeys = e.parseExistingFile(opts.TargetPath)
	}

	sortedKeys := sortKeys(secrets)

	for _, key := range sortedKeys {
		value := secrets[key]

		if existingValue, exists := existingKeys[key]; exists {
			if existingValue == value && !opts.Force {
				preview.SkippedKeys = append(preview.SkippedKeys, key)
			} else if opts.Force {
				preview.UpdatedKeys = append(preview.UpdatedKeys, key)
			} else {
				preview.SkippedKeys = append(preview.SkippedKeys, key)
			}
		} else {
			preview.NewKeys = append(preview.NewKeys, key)
		}
	}

	preview.Content = e.buildContent(secrets, preview, opts)
	return preview, nil
}

func (e *EnvExporter) buildContent(secrets map[string]string, preview *Preview, opts ExportOptions) string {
	var content strings.Builder

	if opts.Append && fileExists(opts.TargetPath) {
		data, err := os.ReadFile(opts.TargetPath)
		if err == nil {
			content.Write(data)
			if !strings.HasSuffix(content.String(), "\n") {
				content.WriteString("\n")
			}
		}
	}

	sortedKeys := sortKeys(secrets)

	if len(preview.NewKeys) > 0 && opts.Append {
		content.WriteString(fmt.Sprintf("\n# Added by veil on %s\n", time.Now().Format("2006-01-02T15:04:05Z")))
	}

	for _, key := range sortedKeys {
		if slices.Contains(preview.NewKeys, key) {
			content.WriteString(fmt.Sprintf("%s=%s\n", key, e.escapeValue(secrets[key])))
		}
	}

	return content.String()
}

func (e *EnvExporter) writeNewFile(preview *Preview, opts ExportOptions) error {
	content := e.buildNewFileContent(preview)
	return safeWriteFile(opts.TargetPath, []byte(content), 0600, opts.Backup, opts.BackupDir)
}

func (e *EnvExporter) buildNewFileContent(preview *Preview) string {
	var content strings.Builder

	allKeys := append(preview.NewKeys, preview.UpdatedKeys...)
	sortedKeys := sortKeysFromSlice(allKeys)

	for _, key := range sortedKeys {
		var value string
		for _, lines := range [][]string{preview.NewKeys, preview.UpdatedKeys} {
			for _, k := range lines {
				if k == key {
					value = preview.Content[strings.Index(preview.Content, key+"=")+len(key)+1:]
					if idx := strings.Index(value, "\n"); idx != -1 {
						value = value[:idx]
					}
					break
				}
			}
		}
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return content.String()
}

func (e *EnvExporter) appendToFile(secrets map[string]string, preview *Preview, opts ExportOptions) error {
	if len(preview.NewKeys) == 0 && len(preview.UpdatedKeys) == 0 {
		return nil
	}

	lines := e.parseFileWithStructure(opts.TargetPath)

	existingKeys := make(map[int]bool)
	for i, line := range lines {
		if !line.isComment && !line.isEmpty && line.key != "" {
			existingKeys[i] = true
		}
	}

	if opts.Force && len(preview.UpdatedKeys) > 0 {
		for i := range lines {
			if slices.Contains(preview.UpdatedKeys, lines[i].key) {
				lines[i].value = secrets[lines[i].key]
				lines[i].original = fmt.Sprintf("%s=%s", lines[i].key, e.escapeValue(lines[i].value))
			}
		}
	}

	if len(preview.NewKeys) > 0 {
		if len(lines) > 0 && !lines[len(lines)-1].isEmpty {
			lines = append(lines, envLine{isEmpty: true, original: ""})
		}
		lines = append(lines, envLine{
			isComment: true,
			original:  fmt.Sprintf("# Added by veil on %s", time.Now().Format("2006-01-02T15:04:05Z")),
		})

		for _, key := range preview.NewKeys {
			lines = append(lines, envLine{
				key:      key,
				value:    secrets[key],
				original: fmt.Sprintf("%s=%s", key, e.escapeValue(secrets[key])),
			})
		}
	}

	var content strings.Builder
	for i, line := range lines {
		content.WriteString(line.original)
		if i < len(lines)-1 || !strings.HasSuffix(line.original, "\n") {
			content.WriteString("\n")
		}
	}

	return safeWriteFile(opts.TargetPath, []byte(content.String()), 0600, opts.Backup, opts.BackupDir)
}

func (e *EnvExporter) parseFileWithStructure(path string) []envLine {
	var lines []envLine
	data, err := os.ReadFile(path)
	if err != nil {
		return lines
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		line := envLine{original: text}

		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			line.isEmpty = true
		} else if strings.HasPrefix(trimmed, "#") {
			line.isComment = true
		} else if key, value, found := strings.Cut(text, "="); found {
			line.key = key
			line.value = e.unescapeValue(value)
		}

		lines = append(lines, line)
	}

	return lines
}

func (e *EnvExporter) parseExistingFile(path string) map[string]string {
	result := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if key, value, found := strings.Cut(line, "="); found {
			result[key] = e.unescapeValue(value)
		}
	}

	return result
}

func (e *EnvExporter) escapeValue(value string) string {
	if strings.Contains(value, " ") || strings.Contains(value, "\t") || strings.Contains(value, "#") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	return value
}

func (e *EnvExporter) unescapeValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, "\\\"", "\"")
	}
	return value
}

func sortKeysFromSlice(keys []string) []string {
	if len(keys) == 0 {
		return nil
	}
	result := slices.Clone(keys)
	slices.Sort(result)
	return slices.Compact(result)
}
