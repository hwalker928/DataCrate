package functions

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func GetDriveLetters() []string {
	// Call the GetLogicalDrives function to get a bitmask of all drive letters
	drives, err := windows.GetLogicalDrives()
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	// Convert the bitmask to a list of drive letters
	var letters []string
	for i := uint(0); i < 26; i++ {
		if (drives>>i)&1 == 1 {
			letter := string('A' + i)
			letters = append(letters, letter+":\\")
		}
	}
	return letters
}
