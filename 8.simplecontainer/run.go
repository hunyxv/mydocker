package simplecontainer

import (
	"mydocker/8.simplecontainer/container"
	"os"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, command string) {
	parent := container.NewParentprocess(tty, command)
	if err := parent.Start(); err != nil {
		logrus.WithError(err).Error("......")
	}

	parent.Wait()
	os.Exit(-1)
}
