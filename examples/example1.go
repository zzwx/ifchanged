package main

import (
	"fmt"
	"log"

	"github.com/zzwx/ifchanged"
)

func main() {
	fileName := "./example1.go"
	err := ifchanged.UsingFile(fileName, fileName+".sha256", func() error {
		fmt.Printf("File %s has changed\n", fileName)
		return nil
	})
	if err != nil {
		log.Fatalf("%+v", err)
	}
	err = ifchanged.UsingFile(fileName, fileName+".sha256", func() error {
		fmt.Printf("This shouldn't show because file %s has been just checked for changes\n", fileName)
		return nil
	})
	if err != nil {
		log.Fatalf("%+v", err)
	}

}
