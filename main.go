package main

import (
	"fmt"
	"os"
	"github.com/davecgh/go-spew/spew"
	"github.com/marxarelli/blubber/config"
)

func main() {
	config, err := config.ReadConfigFile("./blubber.json")

	if err != nil {
		fmt.Println("Error reading config:\n", err)
		os.Exit(1)
	} 

	spew.Dump(config)
}
