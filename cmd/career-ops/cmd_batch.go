package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Run batch evaluation processing",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("batch: not implemented yet")
		return nil
	},
}
