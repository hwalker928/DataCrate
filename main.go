package main

import (
	"archive/zip"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/fatih/color"
	"github.com/hwalker928/DataCrate/functions"
	"github.com/inancgumus/screen"
	"github.com/sqweek/dialog"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func listFiles(root []string, includedExts []string, excludedDirectories []string, excludedFiles []string) []string {
	var files []string
	for _, r := range root {
		filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if functions.ContainsAnyString(strings.ToLower(path), excludedDirectories) {
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

					if functions.ContainsAnyString(fileName, excludedFiles) {
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

func CreateACrate() {
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
	answers := functions.Checkboxes(
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

	zipFiles(filename+"-temp", files)

	end := time.Now()
	elapsed := end.Sub(start)

	functions.EncryptFile(filename+"-temp", "passwordpassword", filename)

	err = os.Remove(filename + "-temp")
	if err != nil {
		fmt.Println(err)
		return
	}
	color.Green("Successfully created a new crate: %s with %d %s in %s.", filename, len(files), english.PluralWord(len(files), "file", "files"), elapsed)
}

func OpenACrate() {
	filename, err := dialog.File().Filter("DataCrate archives", "crate").Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	zipFilename, err := dialog.File().Filter("Zip archives", "zip").Title("Extracted crate destination").SetStartFile(filename + ".zip").Save()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !strings.HasSuffix(zipFilename, ".zip") {
		zipFilename += ".zip"
	}

	password := ""
	prompt := &survey.Password{
		Message: "Please enter the password to decrypt the crate:",
	}
	survey.AskOne(prompt, &password)

	functions.DecryptFile(filename, password, zipFilename)
	if functions.IsValidZipFile(zipFilename) {
		color.Green("Successfully decrypted crate: %s", filename)
		color.Green("Decrypted crate to: %s", zipFilename)
	} else {
		color.Red("Failed to decrypt crate: %s", filename)
		color.Red("Please check that you have entered the correct password, and that the crate is not corrupted.")
		err := os.Remove(zipFilename)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func clearScreen() {
	screen.Clear()
	screen.MoveTopLeft()
}

func main() {
	clearScreen()
	color.Green("\n8888888b.           888              .d8888b.                  888                     \n888  \"Y88b          888             d88P  Y88b                 888                     \n888    888          888             888    888                 888                     \n888    888  8888b.  888888  8888b.  888        888d888 8888b.  888888 .d88b.  .d8888b  \n888    888     \"88b 888        \"88b 888        888P\"      \"88b 888   d8P  Y8b 88K      \n888    888 .d888888 888    .d888888 888    888 888    .d888888 888   88888888 \"Y8888b. \n888  .d88P 888  888 Y88b.  888  888 Y88b  d88P 888    888  888 Y88b. Y8b.          X88 \n8888888P\"  \"Y888888  \"Y888 \"Y888888  \"Y8888P\"  888    \"Y888888  \"Y888 \"Y8888   88888P'\n ")

	function := ""
	prompt := &survey.Select{
		Message: "Select a function:",
		Options: []string{"Create a crate", "Open a crate", "About DataCrates", "Shutdown computer", "Restart computer"},
	}
	survey.AskOne(prompt, &function)

	if function == "Create a crate" {
		CreateACrate()
	} else if function == "Open a crate" {
		OpenACrate()
	} else if function == "About DataCrates" {
		fmt.Println("DataCrates are a new way to backup your data. They are a zip file that contains all of your files, but they are also a self-contained archive that can be opened and browsed like a folder. DataCrates are also encrypted, so you can be sure that your data is safe.")
	} else if function == "Shutdown computer" {
		fmt.Println("Shutting down computer...")
	} else if function == "Restart computer" {
		fmt.Println("Restarting computer...")
	} else {
		fmt.Println("Invalid function selected. Exiting...")
		return
	}
}
