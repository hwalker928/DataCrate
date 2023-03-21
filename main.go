package main

import (
	"archive/zip"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/hwalker928/DataCrate/functions"
	"github.com/sqweek/dialog"
	"io"
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

func listFiles(root []string, includedExts []string, excludedDirectories []string, excludedFiles []string) []string {
	var files []string
	for _, r := range root {
		filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
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
					fileName := strings.ToLower(filepath.Base(path))

					if containsAnyString(fileName, excludedFiles) {
						fmt.Printf("Skipped file %s\n", path)
						break
					}

					files = append(files, path)
					fmt.Printf("Added file %s (size: %s)\n", path, humanize.Bytes(uint64(info.Size())))
					break
				}
			}
			return nil
		})
		fmt.Println("Finished walking path", r)
	}

	return files
}

func zipFiles(zipName string, files []string) {
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
			fmt.Printf("Error opening file %s: %v", file, err.Error())
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

func Checkboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	survey.AskOne(prompt, &res)

	return res
}

func main() {
	filename, err := dialog.File().Filter("DataCrate archives", "crate").Title("Crate destination").SetStartFile("MyCrate-" + time.Now().Format("2006-01-02_15-04-05") + ".crate").Save()
	if err != nil {
		fmt.Println(err)
		return
	}

	if !strings.HasSuffix(filename, ".crate") {
		filename += ".crate"
	}

	fmt.Println("Saving archive as:", filename)

	drives := functions.GetDriveLetters()
	answers := Checkboxes(
		"Select drives to backup:",
		drives,
	)
	s := strings.Join(answers, ", ")
	fmt.Println("Selected drives for backup:", s)
	if len(s) == 0 {
		fmt.Println("No drives selected for backup. Exiting...")
		return
	}

	start := time.Now()

	includedExts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".rtf", ".odt", ".ods", ".odp", ".csv", ".tsv", ".html", ".xml", ".json", ".yaml", ".md", ".tex", ".cfg", ".conf", ".properties", ".prefs", ".plist", ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".cue", ".bin", ".dat", ".db", ".sqlite", ".dbf", ".mdb", ".accdb", ".sql", ".tab", ".tsv", ".dbf", ".dif", ".jpg", ".jpeg", ".png", ".heic", ".gif", ".bmp", ".raw", ".mp4", ".avi", ".wmv", ".mov", ".mkv", ".mp3", ".wav", ".flac", ".aac", ".ogg"}
	excludedDirectories := []string{"temp", "windows", "node_modules", "program files", "programdata", "microsoftteams", "perflogs", "$recycle.bin", "system volume information", "c:\recovery", "cachestorage", "appdata\\local\\packages"}
	excludedFiles := []string{"desktop.ini", "thumbs.db", "ntuser.dat"}

	files := listFiles(answers, includedExts, excludedDirectories, excludedFiles)

	zipFiles(filename, files)

	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Printf("Created zip file %s with %d files.\n", filename, len(files))
	fmt.Printf("Took %s to run\n", elapsed)
}
