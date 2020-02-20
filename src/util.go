package src

import (
	"fmt"
	"log"
	"os"
)

//Check Checks if an error is nil. Prints the error and exits if it isnt't.
func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//CheckWithMessage Checks if an error is nil. Prints a message and exits if it isn't.
func CheckWithMessage(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		os.Exit(1)
	}
}

//CheckCouldntParse Checks if an error is nil. Displays the "couldn't parse" message with additional info
func CheckCouldntParse(err error, info string) {
	if err != nil {
		fmt.Println("❌  Couldn't parse config! " + info)
		os.Exit(1)
	}
}

//Contains check if array contains repo
func Contains(arr []Repo, item Repo) bool {
	for _, a := range arr {
		if a.User == item.User && a.Branch == item.Branch && a.Name == item.Name && a.Tag == item.Tag {
			return true
		}
	}
	return false
}

//IndexOf repo in array
func IndexOf(arr []Repo, item Repo) int {
	for i, a := range arr {
		if a.User == item.User && a.Branch == item.Branch && a.Name == item.Name && a.Tag == item.Tag {
			return i
		}
	}
	return -1
}

//Mapkeys get keys by value
func Mapkeys(m map[string]string, value string) []string {
	var keys []string
	for k, v := range m {
		if v == value {
			keys = append(keys, k)
		}
	}

	return keys
}
