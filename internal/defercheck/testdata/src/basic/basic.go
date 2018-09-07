package basic

import (
	"fmt"
	"log"
)

func BadDefer() error {
	var err error
	defer log.Printf("Error: %v", err) // want "variable err evaluated by defer, then reassigned later"

	err = fmt.Errorf("an error")
	return err
}
