package main

import (
	"fmt"
	"os"

	"github.com/mehmet-f-dogan/fairbook/engine"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: replay <event-log>")
		os.Exit(1)
	}

	_, err := engine.Replay(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println("replay completed successfully")
}
