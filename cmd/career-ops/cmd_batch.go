package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Run batch evaluation processing",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("batch: not implemented yet")
		return nil
	},
}
