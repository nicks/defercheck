package basic

import (
	"fmt"
	"log"
)

func BadDeferBeforeAssign() error {
	var err error
	defer log.Printf("Error: %v", err) // want "variable err evaluated by defer, then reassigned later"

	err = fmt.Errorf("an error")
	return err
}

func BadDeferBeforeReturn() (err error) {
	defer log.Printf("Error: %v", err) // want "variable err evaluated by defer, then returned later"
	return fmt.Errorf("an error")
}
