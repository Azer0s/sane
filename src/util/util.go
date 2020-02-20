package util

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
