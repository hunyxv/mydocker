package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
	subsystemCgroupPath string
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(containerid string, res *ResourceConfig) error {
	if res.MemoryLimit == "" {
		return nil
	}

	subsystemCgroupPath, err := GetCgrouppath(s.Name(), containerid, true)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup memory fail, %+v", err)
	}
	s.subsystemCgroupPath = subsystemCgroupPath
	return nil
}

func (s *MemorySubsystem) Remove() error {
	if s.subsystemCgroupPath == "" {
		return nil
	}
	return os.RemoveAll(s.subsystemCgroupPath)
}

func (s *MemorySubsystem) Apply(containerid string, pid int) error {
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
