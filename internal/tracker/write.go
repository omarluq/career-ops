package tracker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/model"
)

// WriteAtomic writes content to a file atomically via temp file + rename.
func WriteAtomic(path string, content []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".career-ops-*.tmp")
	if err != nil {
		return oops.Wrapf(err, "creating temp file")
	}
	tmpPath := tmp.Name()

	if _, writeErr := tmp.Write(content); writeErr != nil {
		closeErr := tmp.Close()
		removeErr := os.Remove(tmpPath)
		return oops.Wrapf(writeErr, "writing temp file (close=%v, remove=%v)", closeErr, removeErr)
	}
	if err := tmp.Close(); err != nil {
		removeErr := os.Remove(tmpPath)
		return oops.Wrapf(err, "closing temp file (remove=%v)", removeErr)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		removeErr := os.Remove(tmpPath)
		return oops.Wrapf(err, "renaming temp file (remove=%v)", removeErr)
	}
	return nil
}

// BackupAndWrite creates a .bak backup then writes atomically.
func BackupAndWrite(path string, content []byte) error {
	if _, err := os.Stat(path); err == nil {
		if err := writeBackup(path); err != nil {
			return err
		}
	}
	return WriteAtomic(path, content)
}

// writeBackup creates a .bak copy of the file at the given path.
func writeBackup(path string) error {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return oops.Wrapf(err, "reading for backup")
	}
	bakPath := filepath.Clean(cleanPath + ".bak")
	return WriteAtomic(bakPath, data)
}

// UpdateStatus updates a single application's status in applications.md.
func UpdateStatus(
	careerOpsPath string,
	app *model.CareerApplication,
	newStatus string,
) error {
	filePath, err := FindAppsFile(careerOpsPath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return oops.Wrapf(err, "reading %s", filePath)
	}

	lines := strings.Split(string(content), "\n")
	found := false

	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		if app.ReportNumber != "" &&
			strings.Contains(line, fmt.Sprintf("[%s]", app.ReportNumber)) {
			lines[i] = strings.Replace(line, app.Status, newStatus, 1)
			found = true
			break
		}
	}

	if !found {
		return oops.Wrapf(nil,
			"application not found: report %s",
			app.ReportNumber,
		)
	}

	return BackupAndWrite(filePath, []byte(strings.Join(lines, "\n")))
}

// FormatTableLine formats an application as a markdown table row.
func FormatTableLine(
	num int,
	date, company, role, score, status, pdf, report, notes string,
) string {
	return fmt.Sprintf(
		"| %d | %s | %s | %s | %s | %s | %s | %s | %s |",
		num, date, company, role, score, status, pdf, report, notes,
	)
}
