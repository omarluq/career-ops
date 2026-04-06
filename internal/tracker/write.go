package tracker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omarluq/career-ops/internal/model"
)

// WriteAtomic writes content to a file atomically via temp file + rename.
func WriteAtomic(path string, content []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".career-ops-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// BackupAndWrite creates a .bak backup then writes atomically.
func BackupAndWrite(path string, content []byte) error {
	if _, err := os.Stat(path); err == nil {
		bakPath := path + ".bak"
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading for backup: %w", err)
		}
		if err := os.WriteFile(bakPath, data, 0644); err != nil {
			return fmt.Errorf("writing backup: %w", err)
		}
	}
	return WriteAtomic(path, content)
}

// UpdateStatus updates a single application's status in applications.md.
func UpdateStatus(careerOpsPath string, app model.CareerApplication, newStatus string) error {
	filePath, err := FindAppsFile(careerOpsPath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	found := false

	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		if app.ReportNumber != "" && strings.Contains(line, fmt.Sprintf("[%s]", app.ReportNumber)) {
			lines[i] = strings.Replace(line, app.Status, newStatus, 1)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("application not found: report %s", app.ReportNumber)
	}

	return BackupAndWrite(filePath, []byte(strings.Join(lines, "\n")))
}

// FormatTableLine formats an application as a markdown table row.
func FormatTableLine(num int, date, company, role, score, status, pdf, report, notes string) string {
	return fmt.Sprintf("| %d | %s | %s | %s | %s | %s | %s | %s | %s |",
		num, date, company, role, score, status, pdf, report, notes)
}
