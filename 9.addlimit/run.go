package addlimit

import (
	"mydocker/8.simplecontainer/container"
	"mydocker/9.addlimit/cgroup"
	"mydocker/9.addlimit/cgroup/subsystems"
	"strings"

	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

func Run(tty bool, res *subsystems.ResourceConfig, command string) {
	parent := container.NewParentprocess(tty, command)
	if err := parent.Start(); err != nil {
		logrus.WithError(err).Error("......")
	}

	containerid := strings.ReplaceAll(uuid.NewRandom().String(), "-", "")
	cgroupManager := cgroup.NewCgroupManager(containerid, res)
	defer cgroupManager.Destroy()

	cgroupManager.Set()
	cgroupManager.Apply(parent.Process.Pid)

	parent.Wait()
}
