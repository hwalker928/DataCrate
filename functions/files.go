package functions

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// ListFiles recursively walks through directories starting from the root paths specified and returns a list of all files that match the given criteria
func ListFiles(root []string, includedExts []string, excludedDirectories []string, excludedFiles []string) []string {
	var files []string
	for _, r := range root {
		// Walk through the file path and execute the provided function for each file/directory found
		filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
			// If it's a directory and is excluded, skip walking its contents
			if info.IsDir() {
				if ContainsAnyString(strings.ToLower(path), excludedDirectories) {
					return filepath.SkipDir
				}
				return nil
			}
			// If there's an error accessing the file, log it and continue
			if err != nil {
				fmt.Printf("Error walking path %s: %v\n", path, err.Error())
				return nil
			}
			// Check if the file's extension is one of the included extensions
			ext := strings.ToLower(filepath.Ext(path))
			for _, includedExt := range includedExts {
				if ext == includedExt {
					fileName := strings.ToLower(filepath.Base(path))
					// Check if the file's name is one of the excluded files
					if ContainsAnyString(fileName, excludedFiles) {
						break
					}
					// Append the file path to the list of files
					files = append(files, path)
					break
				}
			}
			return nil
		})
		// Log that we finished walking the current path
		fmt.Println("Finished walking path", r)
	}

	// Return the list of matching files
	return files
}

func ZipFiles(zipName string, files []string) {
	fmt.Println("Zipping files...")
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Printf("Error creating zip: %s", err.Error())
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		fileToZip, err := os.Open(file)
		if err != nil {
			color.Red("Error opening file %s", file)
			return
		}
		defer fileToZip.Close()

		// Get the file information
		info, err := fileToZip.Stat()
		if err != nil {
			fmt.Printf("Error getting file info of %s: %v", file, err.Error())
			return
		}

		// Create a new file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			fmt.Printf("Error creating file header for %s: %v", file, err.Error())
			return
		}

		// Set the file name to preserve directory structure
		header.Name = strings.ReplaceAll(filepath.ToSlash(file), ":", "")

		// Add the file header to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fmt.Printf("Error creating file header for %s: %v", file, err.Error())
			return
		}

		// Write the file content to the zip archive
		if _, err := io.Copy(writer, fileToZip); err != nil {
			fmt.Printf("Error copying file %s to zip: %v", file, err.Error())
			return
		}
	}

	return
}
