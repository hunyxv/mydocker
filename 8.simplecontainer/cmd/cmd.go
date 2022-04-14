package main

import (
	"os"

	simplecontainer "mydocker/8.simplecontainer"
	"mydocker/8.simplecontainer/container"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)
var (
	tty bool
)

func init() {
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")

	rootCmd.AddCommand(runCommand)
	rootCmd.AddCommand(initCommand)
}

var rootCmd = &cobra.Command{
	Use:   "mydocker",
	Short: "mydocker is a simple container runtime implementation .",
	Long: `The purpose of this project is to learn how docker works and how to write a docker by ourselves.
	Enjoy it, just for fun.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
	},
}

var runCommand = &cobra.Command{
	Use: "run [-it tty]",
	Short: "Create a container with namespace and cgroups limit mydocker run -it [command]",
	Run: func(cmd *cobra.Command, args []string) {
		simplecontainer.Run(tty, args[0])
	},
}

var initCommand = &cobra.Command{
	Use: "init",
	Short: "I nit container process run user's process in container . Do not call it outside",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			return
		}

		container.RunContainerInitProcess(args[0], args[1:])
	},
}

func main(){
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}