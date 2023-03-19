package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func containsAnyString(str string, substr []string) bool {
	for _, s := range substr {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}

func listFiles(root string, includedExts []string, excludedDirectories []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if containsAnyString(strings.ToLower(path), excludedDirectories) {
				return filepath.SkipDir
			}
			return nil
		}
		if err != nil {
			fmt.Printf("Error walking path %s: %v\n", path, err.Error())
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, includedExt := range includedExts {
			if ext == includedExt {
				files = append(files, path)
				fmt.Printf("Added file %s\n", path)
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func zipFiles(zipName string, files []string) error {
	zipFile, err := os.Create(zipName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		fileToZip, err := os.Open(file)
		if err != nil {
			fmt.Printf("Error opening file %s: %v", file, err.Error())
			return nil
		}
		defer fileToZip.Close()

		// Get the file information
		info, err := fileToZip.Stat()
		if err != nil {
			fmt.Printf("Error getting file info of %s: %v", file, err.Error())
			return nil
		}

		// Create a new file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			fmt.Printf("Error creating file header for %s: %v", file, err.Error())
			return nil
		}

		// Set the file name to preserve directory structure
		header.Name = strings.ReplaceAll(filepath.ToSlash(file), ":", "")

		// Add the file header to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fmt.Printf("Error creating file header for %s: %v", file, err.Error())
			return nil
		}

		// Write the file content to the zip archive
		if _, err := io.Copy(writer, fileToZip); err != nil {
			fmt.Printf("Error copying file %s to zip: %v", file, err.Error())
			return nil
		}
	}

	return nil
}

func main() {
	// ask user for root directory input
	fmt.Println("Enter the root directory to zip:")
	root := "."
	_, err := fmt.Scanln(&root)
	if err != nil {
		return
	}

	start := time.Now()

	includedExts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".rtf", ".odt", ".ods", ".odp", ".csv", ".tsv", ".html", ".xml", ".json", ".yaml", ".md", ".tex", ".cfg", ".conf", ".properties", ".prefs", ".plist", ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".cue", ".bin", ".dat", ".db", ".sqlite", ".dbf", ".mdb", ".accdb", ".sql", ".tab", ".tsv", ".dbf", ".dif"}
	excludedDirectories := []string{"temp", "windows", "node_modules", "program files", "programdata", "microsoftteams", "perflogs", "$recycle.bin", "system volume information", "c:\recovery", "cachestorage", "appdata\\local\\packages"}

	files, err := listFiles(root, includedExts, excludedDirectories)
	if err != nil {
		log.Fatal(err)
	}

	zipName := "output.zip"
	err = zipFiles(zipName, files)
	if err != nil {
		log.Fatal(err)
	}

	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Printf("Created zip file %s with %d files.\n", zipName, len(files))
	fmt.Printf("Took %s to run\n", elapsed)

	_, err = fmt.Scanln()
	if err != nil {
		return
	}
}
