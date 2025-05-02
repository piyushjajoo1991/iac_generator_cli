package main

import (
	"fmt"
	"os"

	"github.com/riptano/iac_generator_cli/cmd/iacgen"
	"github.com/riptano/iac_generator_cli/examples"
)

func main() {
	// Check if we want to run the example or the CLI
	if len(os.Args) > 1 && os.Args[1] == "example" {
		// Run the complete example with debug output
		fmt.Println("Running complete IaC Generator example with debug output...")
		examples.RunFromCommandLine()
	} else {
		// Run the standard CLI
		iacgen.Execute()
	}
}