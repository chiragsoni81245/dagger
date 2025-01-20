package utils

import (
	"archive/zip"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UpdateTaskStatus(db *sql.DB, taskId int, status string) error {
    result, err := db.Exec(`
    UPDATE task
    SET status=$1
    WHERE id=$2
    `, status, taskId)
    rowsAffected, err := result.RowsAffected()

    if err != nil {
        return err
    }

    if rowsAffected != 1 {
        return errors.New("error in updating task status")
    }

    return nil
}

func UpdateDagStatus(db *sql.DB, dagId int, status string) error {
    result, err := db.Exec(`
    UPDATE dag
    SET status=$1
    WHERE id=$2
    `, status, dagId)
    rowsAffected, err := result.RowsAffected()

    if err != nil {
        return err
    }

    if rowsAffected != 1 {
        return errors.New("error in updating dag status")
    }

    return nil
}


// Unzip extracts a ZIP file to the specified destination directory.
func Unzip(src, dest string) error {
	// Open the ZIP file
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, file := range r.File {
		if err := extractAndWriteFile(file, dest); err != nil {
			return fmt.Errorf("error extracting file %q: %w", file.Name, err)
		}
	}

	return nil
}

// extractAndWriteFile handles extraction of individual files and directories.
func extractAndWriteFile(file *zip.File, dest string) error {
	// Resolve the file path within the destination directory
	targetPath := filepath.Join(dest, file.Name)

	// Prevent ZipSlip by checking the relative path
	relPath, err := filepath.Rel(dest, targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path: %w", err)
	}
	if strings.HasPrefix(relPath, "..") {
		return errors.New("illegal file path: potential ZipSlip attack")
	}

	// Create directory if the file is a directory
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Extract and write file
	return writeFile(file, targetPath)
}

// writeFile extracts a single file from the ZIP archive and writes it to disk.
func writeFile(file *zip.File, targetPath string) error {
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip file content: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

