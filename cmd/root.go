package cmd

import (
	"fmt"
	"os"

	"github.com/sherifabdlnaby/rubban/rubban"
	"github.com/sherifabdlnaby/rubban/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rubban",
	Short: "Start rubban",
	Long: `Start rubban according to configuration loaded from:
	1- Environment Variables.	(ex: RUBBAN_KIBANA_HOST=https://kibana:5601"
	2- .env file.
	3- rubban.(yaml|yml|json|toml)
	(values from the earlier overwrite the latter).`,
	Run: func(cmd *cobra.Command, args []string) {
		rubban.Main()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs	 to happen once to the rootCmd.
func Execute() {
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate(version.Get())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}
