package main

import (
	"fmt"

	"github.com/fatih/color"
)

func main() {
	highlight := color.New(color.FgWhite, color.BgGreen).SprintFunc()
	fmt.Printf("Hello, %s!\n", highlight("World"))
}
