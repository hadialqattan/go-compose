/*
Go-compose starting Cobra sub-command.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/hadialqattan/go-compose/utils"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start go-compose.yaml services",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.GetConfig(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		processor := utils.CreateProcessor(config)
		utils.ShutdownSignalObserver(&processor.Core)
		processor.Core.Run()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
