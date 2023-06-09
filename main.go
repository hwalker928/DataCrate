package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize/english"
	"github.com/fatih/color"
	"github.com/hwalker928/DataCrate/config"
	"github.com/hwalker928/DataCrate/functions"
	"github.com/inancgumus/screen"
	"github.com/sqweek/dialog"
)

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

	encryptionMethod := ""
	prompt := &survey.Select{
		Message: "Select an encryption method:",
		Options: []string{"Key file (Recommended)", "User-defined password", "Random password", "No encryption (Not recommended)"},
	}
	survey.AskOne(prompt, &encryptionMethod)

	password := ""
	showPasswordAtEnd := false

	switch encryptionMethod {
	case "Key file (Recommended)":
		keyFilename, err := dialog.File().Filter("DataCrate key", "key").Title("Crate key destination").SetStartFile(filename + ".key").Save()
		if err != nil {
			fmt.Println(err)
			return
		}

		password = functions.GenerateRandomString(4096)

		// write the key to the key file
		err = ioutil.WriteFile(keyFilename, []byte(password), 0644)
		if err != nil {
			fmt.Println("Error writing key to file:", err.Error())
			return
		}
	case "User-defined password":
		password = ""
		prompt2 := &survey.Password{
			Message: "Please enter the password to encrypt the crate:",
		}
		survey.AskOne(prompt2, &password)
	case "Random password":
		password = functions.GenerateRandomString(32)
		showPasswordAtEnd = true
	case "No encryption (Not recommended)":
		color.Red("WARNING: Crate will not be encrypted. This is not recommended.")
	}

	start := time.Now()

	includedExts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".rtf", ".odt", ".ods", ".odp", ".csv", ".tsv", ".html", ".xml", ".json", ".yaml", ".md", ".tex", ".cfg", ".conf", ".properties", ".prefs", ".plist", ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".cue", ".bin", ".dat", ".db", ".sqlite", ".dbf", ".mdb", ".accdb", ".sql", ".tab", ".tsv", ".dbf", ".dif", ".jpg", ".jpeg", ".png", ".heic", ".gif", ".bmp", ".raw", ".mp4", ".avi", ".wmv", ".mov", ".mkv", ".mp3", ".wav", ".flac", ".aac", ".ogg"}
	excludedDirectories := []string{"temp", "windows", "node_modules", "program files", "programdata", "microsoftteams", "perflogs", "$recycle.bin", "system volume information", "c:\recovery", "cache", "appdata\\local\\packages"}
	excludedFiles := []string{"desktop.ini", "thumbs.db", "ntuser.dat"}

	files := functions.ListFiles(answers, includedExts, excludedDirectories, excludedFiles)

	functions.ZipFiles(filename+"-temp", files)

	end := time.Now()
	elapsed := end.Sub(start)

	functions.EncryptFile(filename+"-temp", password, filename)

	err = os.Remove(filename + "-temp")
	if err != nil {
		fmt.Println(err)
		return
	}

	color.Green("Successfully created a new crate: %s with %d %s in %s.", filename, len(files), english.PluralWord(len(files), "file", "files"), elapsed)

	if showPasswordAtEnd {
		color.Cyan("Random password generated: %s", password)
	}
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

	encryptionMethod := ""
	prompt := &survey.Select{
		Message: "Select the encryption method used for this crate:",
		Options: []string{"Key file", "Password", "No encryption"},
	}
	survey.AskOne(prompt, &encryptionMethod)

	password := ""

	switch encryptionMethod {
	case "Key file":
		keyFilename, err := dialog.File().Filter("DataCrate key", "key").Load()
		if err != nil {
			fmt.Println(err)
			return
		}

		password = functions.ReadKeyFile(keyFilename)

	case "Password":
		prompt := &survey.Password{
			Message: "Please enter the password to decrypt the crate:",
		}
		survey.AskOne(prompt, &password)
	case "No encryption":
		color.Red("WARNING: Crate is not encrypted. This is not recommended.")
	}

	start := time.Now()

	functions.DecryptFile(filename, password, zipFilename)

	end := time.Now()
	elapsed := end.Sub(start)

	if functions.IsValidZipFile(zipFilename) {
		color.Green("Successfully decrypted crate %s in %s.", filename, elapsed)
	} else {
		color.Red("Failed to decrypt crate: %s", filename)
		color.Red("Please check that you have chosen the correct encryption method and entered the correct information, and that the crate is not corrupted.")
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
		Options: []string{"Create a crate", "Open a crate", "About DataCrates", "Exit DataCrates", "Shutdown computer", "Restart computer"},
	}
	survey.AskOne(prompt, &function)

	switch function {
	case "Create a crate":
		CreateACrate()
		fmt.Scanln()
		main()
	case "Open a crate":
		OpenACrate()
		fmt.Scanln()
		main()
	case "Exit DataCrates":
		fmt.Println("Exiting DataCrates... Have a nice day!")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	case "About DataCrates":
		color.Cyan("You are running DataCrates v%s created by %s", config.Version, config.Author)
		fmt.Println("\nDataCrates is a new way to back up your documents.\nCrates are an archive (known as .crate) that contains all of your personal documents, that becomes a zip file when decrypted.\nCrates are encrypted with a user-defined password, so you can be sure that your data is safe")
		fmt.Scanln()
		main()
	case "Shutdown computer":
		color.Red("Shutting down computer...")
		time.Sleep(1 * time.Second)
		err := exec.Command("shutdown", "/s", "/t", "0").Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "Restart computer":
		color.Red("Restarting computer...")
		time.Sleep(1 * time.Second)
		err := exec.Command("shutdown", "/r", "/t", "0").Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
