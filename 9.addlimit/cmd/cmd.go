package main

import (
	"os"

	"mydocker/8.simplecontainer/container"
	addlimit "mydocker/9.addlimit"
	"mydocker/9.addlimit/cgroup/subsystems"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tty      bool
	memory   string
	cpushare string
	cpuset   string
)

func init() {
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
	runCommand.Flags().StringVarP(&memory, "memory", "m", "", "Memory limit")
	runCommand.Flags().StringVarP(&cpushare, "cpu-share", "c", "", "CPU shares (relative weight)")
	runCommand.Flags().StringVarP(&cpuset, "cpuset", "", "", "CPUs in which to allow execution")

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
	Use:   "run [-t tty]",
	Short: "Create a container with namespace and cgroups limit mydocker run -it [command]",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info(memory, cpuset, cpushare)
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: memory,
			CpuSet:      cpuset,
			CpuShare:    cpushare,
		}

		addlimit.Run(tty, resConf, args[0])
	},
}

var initCommand = &cobra.Command{
	Use:    "init",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			return
		}

		container.RunContainerInitProcess(args[0], args[1:])
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
