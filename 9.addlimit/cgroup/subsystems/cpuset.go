package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CpusetSubsystem 在多核机器上设置 cgroup 中进程可以使用的 CPU ;
type CpusetSubsystem struct {
	subsystemCgroupPath string
}

func (s *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubsystem) Set(containerid string, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}

	subsystemCgroupPath, err := GetCgrouppath(s.Name(), containerid, true)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup cpuset.cpus fail, %+v", err)
	}
	err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.mems"), []byte("0"), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup cpuset.cpus fail, %+v", err)
	}
	s.subsystemCgroupPath = subsystemCgroupPath
	return nil
}

func (s *CpusetSubsystem) Remove() error {
	if s.subsystemCgroupPath == "" {
		return nil
	}
	return os.RemoveAll(s.subsystemCgroupPath)
}

func (s *CpusetSubsystem) Apply(containerid string, pid int) error {
	subsystemCgroupPath := s.subsystemCgroupPath
	var err error
	if subsystemCgroupPath == "" {
		subsystemCgroupPath, err = GetCgrouppath(s.Name(), containerid, true)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup task id fail, %+v", err)
	}
	s.subsystemCgroupPath = subsystemCgroupPath
	return nil
}
