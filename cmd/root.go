package cmd

import (
	"fmt"
	"os"

	"github.com/sherifabdlnaby/bosun/bosun"
	"github.com/sherifabdlnaby/bosun/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Bosun",
	Short: "xxx",
	Long:  `XXXXXX`,
	Run: func(cmd *cobra.Command, args []string) {
		bosun.Main()
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
