// Package cmd gather all command line interface functions.
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/perebaj/esaj/esaj"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "esaj",
	Short: "esaj is a tool to collect and download data from ESAJ",
	Long:  `esaj is a tool to collect and download data from ESAJ`,
	Run: func(_ *cobra.Command, _ []string) {
		// Do Stuff Here
	},
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect all data from ESAJ related to a specific OAB number or process",
	Long:  `Collect all data from ESAJ related to a specific OAB number or process`,
	Run: func(cmd *cobra.Command, _ []string) {
		oab, _ := cmd.Flags().GetString("oab")
		processID, _ := cmd.Flags().GetString("process")
		output, _ := cmd.Flags().GetString("output")
		ctx := cmd.Context()
		if oab == "" && processID == "" {
			fmt.Println("Error: You must provide either an OAB number or a process ID")
			_ = cmd.Usage()
			return
		}

		eClient := esaj.New(esaj.Config{}, &http.Client{
			Timeout: 30 * time.Second,
		})

		if oab != "" {
			fmt.Println("Collecting data for OAB number:", oab)
			seed, err := eClient.SearchByOAB(ctx, oab)
			if err != nil {
				fmt.Println("Error searching by OAB:", err)
				return
			}
			bar := progressbar.Default(int64(len(seed)))
			var allProcesses []esaj.ProcessBasicInfo
			for _, s := range seed {
				processBasicInfo, err := eClient.FetchBasicProcessInfo(ctx, s.URL, s.ProcessID)
				if err != nil {
					fmt.Println("Error fetching basic process info:", err)
					return
				}
				processBasicInfo.OAB = oab
				allProcesses = append(allProcesses, *processBasicInfo)
				_ = bar.Add(1)
			}

			data, err := json.Marshal(allProcesses)
			if err != nil {
				fmt.Println("Error marshalling basic process info:", err)
				return
			}

			f, err := os.Create(output)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer func() {
				_ = f.Close()
			}()
			_, err = f.Write(data)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}

		if processID != "" {
			foro, err := esaj.ForoNumeroUnificado(processID)
			if err != nil {
				fmt.Println("Error getting foro:", err)
				return
			}
			fmt.Println("Collecting data for Process ID:", processID)
			processCode, err := eClient.ProcessCodeByProcessID(processID)
			if err != nil {
				fmt.Println("Error getting process code:", err)
				return
			}
			url := fmt.Sprintf("https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=%s&processo.foro=%s", processCode, foro)
			processBasicInfo, err := eClient.FetchBasicProcessInfo(ctx, url, processID)
			if err != nil {
				fmt.Println("Error fetching basic process info:", err)
				return
			}
			fmt.Printf("Basic process info: %+v\n", processBasicInfo)
			resp, err := json.Marshal(processBasicInfo)
			if err != nil {
				fmt.Println("Error marshalling basic process info:", err)
				return
			}
			fmt.Println(string(resp))
		}
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(downloadCmd)
	collectCmd.Flags().StringP("oab", "o", "", "OAB number to search")
	collectCmd.Flags().StringP("process", "p", "", "Process ID to search")
	collectCmd.Flags().StringP("output", "O", "processes.json", "Output file")
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download all PDFs documents related to a specific process",
	Long:  `Download all PDFs documents related to a specific process`,
	Run: func(_ *cobra.Command, _ []string) {
		// Do Stuff Here
	},
}

// Execute is the entry point for the command line interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
