package main

import (
	"fmt"

	"git.bacardi55.io/bacardi55/walgot/cmd"
)

// Main function.
func main() {
	if c, e := cmd.Init(); e != nil {
		fmt.Println("Error loading walgot", e)
	} else {
		c.Run()
	}
}
