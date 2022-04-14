package container

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command %s", command)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		logrus.WithError(err).Error("mount fail")
		return err
	}

	if err = syscall.Exec(command, args, os.Environ()); err != nil {
		logrus.WithError(err).Error("exec err")
		return err
	}
	return nil
}