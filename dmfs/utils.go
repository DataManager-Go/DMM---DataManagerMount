package dmfs

import (
	"fmt"

	libdm "github.com/DataManager-Go/libdatamanager"
)

// Print an response error for normies
func printResponseError(err error, msg string) {
	if err == nil {
		return
	}

	switch err.(type) {
	case *libdm.ResponseErr:
		lrerr := err.(*libdm.ResponseErr)

		var cause string

		if lrerr.Response != nil {
			cause = lrerr.Response.Message
		} else if lrerr.Err != nil {
			cause = lrerr.Err.Error()
		} else {
			cause = lrerr.Error()
		}

		printError(msg, cause)
	default:
		if err != nil {
			printError(msg, err.Error())
		} else {
			printError(msg, "no error provided")
		}
	}
}

func printError(message interface{}, err string) {
	fmt.Println(getError(message, err))
}

func getError(message interface{}, err string) string {
	return fmt.Sprintf("Error %s: %s\n", message, err)
}
