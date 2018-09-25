package deployer

import (
	"log"
)

func CheckError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
