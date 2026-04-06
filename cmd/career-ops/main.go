// career-ops is the CLI entry point for the AI job search pipeline.
package main

import "os"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
