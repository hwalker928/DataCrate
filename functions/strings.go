package functions

import "strings"

func ContainsAnyString(str string, substr []string) bool {
	for _, s := range substr {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}
