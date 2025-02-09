package utils

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

// createTarFromZip reads a zip file and writes its contents into a tarball.
func CreateTarFromZip(zipFilePath string, tarBuf *bytes.Buffer, renameFiles map[string]string) error {
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	stat, err := zipFile.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return err
	}

	tarWriter := tar.NewWriter(tarBuf)
	defer tarWriter.Close()

	for _, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

        fileName := file.Name
        if _, ok := renameFiles[fileName]; ok {
            fileName = renameFiles[fileName]
            delete(renameFiles, fileName)
        }

		header := &tar.Header{
			Name: fileName,
			Mode: 0600,
			Size: int64(file.UncompressedSize64),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tarWriter, fileReader); err != nil {
			return err
		}
	}

    if len(renameFiles) > 0 {
        keys := make([]string, 1)
        for k := range renameFiles {
            keys = append(keys, k)
        }
        return fmt.Errorf("Files not found in zip, %s", strings.Join(keys, ", "))
    }

	return nil
}

