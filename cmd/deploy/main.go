package main

import (
	"fmt"
	"os"
)

const configPath = "configs/deployment.yaml"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: classifaas [deploy|generate]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "deploy":
		runDeploy()
	case "generate":
		runGenerate()
	case "remove":
		runRemove()
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
