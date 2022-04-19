package rootfs

import (
	"mydocker/11.overlayfs/cgroup"
	"mydocker/11.overlayfs/cgroup/subsystems"
	"mydocker/11.overlayfs/container"
	"strings"

	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

func Run(tty bool, res *subsystems.ResourceConfig, command string, image string, rm bool) error {
	parent, err := container.NewParentProcess(tty, command, image, rm)
	if err != nil {
		return err
	}
	defer parent.Release()

	if err := parent.Start(); err != nil {
		logrus.WithError(err).Error("......")
		return err
	}

	pid, _ := parent.PID()

	containerid := strings.ReplaceAll(uuid.NewRandom().String(), "-", "")
	cgroupManager := cgroup.NewCgroupManager(containerid, res)
	defer cgroupManager.Destroy()

	cgroupManager.Set()
	cgroupManager.Apply(pid)

	return parent.Wait()
}
