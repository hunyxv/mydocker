package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// 设置进程cpu使用率
type CpuSubsystem struct {
	subsystemCgroupPath string
}

func (s *CpuSubsystem) Name() string {
	return "cpu"
}

func (s *CpuSubsystem) Set(containerid string, res *ResourceConfig) error {
	if res.CpuShare == "" {
		return nil
	}

	subsystemCgroupPath, err := GetCgrouppath(s.Name(), containerid, true)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup cpu.share fail, %+v", err)
	}
	s.subsystemCgroupPath = subsystemCgroupPath
	return nil
}

func (s *CpuSubsystem) Remove() error {
	if s.subsystemCgroupPath == "" {
		return nil
	}
	return os.RemoveAll(s.subsystemCgroupPath)
}

func (s *CpuSubsystem) Apply(containerid string, pid int) error {
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
