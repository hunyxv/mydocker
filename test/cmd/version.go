package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Verbose bool
var Source string

func init() {
	versionCmd.Flags().StringVarP(&Source, "source", "s", "", "Source directory to read from")
	rootCmd.AddCommand(versionCmd)

	cmdTimes.Flags().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

	rootCmd.AddCommand(cmdPrint, cmdEcho)
	cmdEcho.AddCommand(cmdTimes)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
		s := cmd.Flag("source")
		fmt.Println(s.Value.String())
	},
}
