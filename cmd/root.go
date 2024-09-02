// Package cmd gather all command line interface functions.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "esaj-collector",
	Short: "Extract unstructured data from PDFs and transform it into structured data",
	Long:  `Use AI to extract relevant information from the data`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(downloadCmd)
	collectCmd.Flags().StringP("oab", "o", "", "OAB number to search")
	collectCmd.Flags().StringP("process", "p", "", "Process ID to search")
	downloadCmd.Flags().StringP("process", "p", "", "Process ID to search")
	downloadCmd.Flags().StringP("output", "o", "", "Output folder to save the PDFs")
	downloadCmd.Flags().BoolP("markdown", "m", false, "Convert PDFs to markdown")
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect all data from ESAJ related to a specific OAB number or process",
	Long:  `Collect all data from ESAJ related to a specific OAB number or process`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download all PDFs documents related to a specific process",
	Long:  `Download all PDFs documents related to a specific process`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
