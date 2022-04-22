package detach

import (
	"encoding/json"
	"mydocker/12.detach/cgroup"
	"mydocker/12.detach/cgroup/subsystems"
	"mydocker/12.detach/container"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

var (
	DefaultInfoLocation string = "/var/lib/mydocker/containers"
)

type RunOptions struct {
	TTY        bool
	AuthRemove bool
	Memory     string
	Cpushare   string
	Cpuset     string
	Image      string
	Volumes    []string
	Detach     bool

	AllArgs []string
}

func Run(command []string, opts *RunOptions) error {
	if opts.Detach {
		args := []string{"run"}
		args = append(args, opts.AllArgs...)
		logrus.Infof("%+v, %d", args, len(args))
		cmd := exec.Command("/proc/self/exe", args...)
		return cmd.Start()
	}

	resConf := &subsystems.ResourceConfig{
		MemoryLimit: opts.Memory,
		CpuSet:      opts.Cpuset,
		CpuShare:    opts.Cpushare,
	}
	if opts.Memory != "" {
		subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.MemorySubsystem{})
	}
	if opts.Cpuset != "" {
		subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.CpusetSubsystem{})
	}
	if opts.Cpushare != "" {
		subsystems.SubsystemsIns = append(subsystems.SubsystemsIns, &subsystems.CpuSubsystem{})
	}

	parent, err := container.NewParentProcess(opts.TTY, command, opts.Image, opts.AuthRemove, opts.Volumes)
	if err != nil {
		return err
	}
	defer parent.Release()

	if err := parent.Start(); err != nil {
		logrus.WithError(err).Error("......")
		return err
	}

	pid, _ := parent.PID()

	containerInfo := &container.ContainerInfo{
		Pid:         pid,
		Id:          parent.ContainerID(),
		Name:        parent.ContainerID(),
		Command:     strings.Join(command, " "),
		CreatedTime: time.Now().String(),
		Volume:      opts.Volumes,
	}

	saveContainerInfo(DefaultInfoLocation, containerInfo)

	containerid := strings.ReplaceAll(uuid.NewRandom().String(), "-", "")
	cgroupManager := cgroup.NewCgroupManager(containerid, resConf)
	cgroupManager.Set()
	cgroupManager.Apply(pid)

	defer func() {
		cgroupManager.Destroy()
		removeContainerInfo(DefaultInfoLocation + "/" + containerInfo.Id)
	}()

	return parent.Wait()
}

func saveContainerInfo(path string, info *container.ContainerInfo) error {
	exist, err := container.PathExists(path)
	if err != nil {
		return err
	}

	if !exist {
		if err := os.MkdirAll(path, 0710); err != nil {
			return err
		}
	}

	infoByte, _ := json.Marshal(info)
	return os.WriteFile(path+"/"+info.Id+".json", infoByte, 0644)
}

func removeContainerInfo(path string) {
	os.RemoveAll(path + ".json")
}
