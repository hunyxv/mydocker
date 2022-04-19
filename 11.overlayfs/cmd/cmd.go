package main

import (
	"os"

	overlayfs "mydocker/11.overlayfs"
	"mydocker/11.overlayfs/cgroup/subsystems"
	"mydocker/11.overlayfs/container"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tty      bool
	rm       bool
	memory   string
	cpushare string
	cpuset   string
	image    string
)

func init() {
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
	runCommand.Flags().BoolVarP(&rm, "rm", "", false, "Automatically remove the container when it exits")
	runCommand.Flags().StringVarP(&memory, "memory", "m", "", "Memory limit")
	runCommand.Flags().StringVarP(&cpushare, "cpu-share", "c", "", "CPU shares (relative weight)")
	runCommand.Flags().StringVarP(&cpuset, "cpuset", "", "", "CPUs in which to allow execution")
	runCommand.Flags().StringVarP(&image, "image", "i", "", "Base image")

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
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: memory,
			CpuSet:      cpuset,
			CpuShare:    cpushare,
		}
		if memory != "" {
			subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.MemorySubsystem{})
		}
		if cpuset != "" {
			subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.CpusetSubsystem{})
		}
		if cpushare != "" {
			subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.CpuSubsystem{})
		}

		err := overlayfs.Run(tty, resConf, args[0], image, rm)
		if err != nil {
			logrus.WithError(err).Error("run error")
		}
	},
}

var initCommand = &cobra.Command{
	Use:    "init",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("args: ", args, len(args))
		if len(args) == 0 {
			return
		}

		err := container.RunContainerInitProcess(args[0], args[1:])
		if err != nil {
			logrus.WithError(err).Error("..")
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
