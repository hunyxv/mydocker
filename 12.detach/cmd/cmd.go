package main

import (
	"encoding/json"
	"os"

	mydocker "mydocker/12.detach"
	"mydocker/12.detach/container"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	tty      bool
	rm       bool
	memory   string
	cpushare string
	cpuset   string
	image    string
	volumes  []string
	detach   bool
)

func init() {
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
	runCommand.Flags().BoolVarP(&rm, "rm", "", false, "Automatically remove the container when it exits")
	runCommand.Flags().StringVarP(&memory, "memory", "m", "", "Memory limit")
	runCommand.Flags().StringVarP(&cpushare, "cpu-share", "", "", "CPU shares (relative weight)")
	runCommand.Flags().StringVarP(&cpuset, "cpuset", "", "", "CPUs in which to allow execution")
	runCommand.Flags().StringVarP(&image, "image", "i", "", "Base image")
	runCommand.Flags().StringArrayVarP(&volumes, "volume", "v", []string{}, "Bind mount a volume")
	runCommand.Flags().BoolVarP(&detach, "detach", "d", false, "Run container in background and print container ID")

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
	Use:                "run [-t tty]",
	Short:              "Create a container with namespace and cgroups limit mydocker run -it [command]",
	Args:               cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		args = append([]string{"--"}, args...)
		var runArgs []string
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "detach" || f.Name == "tty" || f.Name == "help" {
				return
			}
			if f.Value.Type() == "stringArray" {
				var value []string
				json.Unmarshal([]byte(f.Value.String()), &value)
				for _, i := range value {
					runArgs = append(runArgs, "--"+f.Name, i)
				}
				return
			}

			if f.Value.Type() == "bool" {
				runArgs = append(runArgs, "--"+f.Name)
				return
			}

			if f.Value.String() != "" {
				runArgs = append(runArgs, "--"+f.Name, f.Value.String())
			}
		})
		runArgs = append(runArgs, args...)

		err := mydocker.Run(args, &mydocker.RunOptions{
			TTY:        tty,
			AuthRemove: rm,
			Memory:     memory,
			Cpushare:   cpushare,
			Cpuset:     cpuset,
			Image:      image,
			Volumes:    volumes,
			Detach:     detach,

			AllArgs: runArgs,
		})
		if err != nil {
			logrus.WithError(err).Error("run error")
		}
	},
}

var initCommand = &cobra.Command{
	Use:                "init",
	Hidden:             true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("args: ", args, len(args))
		if len(args) == 0 {
			return
		}

		err := container.RunContainerInitProcess(args[0], args)
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
