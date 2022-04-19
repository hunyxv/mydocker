package cgroup

import (
	"mydocker/10.rootfs/cgroup/subsystems"

	"github.com/sirupsen/logrus"
)

type CgroupManager struct {
	ContainerID string
	Resource    *subsystems.ResourceConfig
}

func NewCgroupManager(containerid string, resource *subsystems.ResourceConfig) *CgroupManager {
	return &CgroupManager{
		ContainerID: containerid,
		Resource:    resource,
	}
}

// 将进程pid加入到这个cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		err := subSysIns.Apply(c.ContainerID, pid)
		if err != nil {
			logrus.WithError(err).Error("cgroup subsystem apply fail")
		}
	}
	return nil
}

// 设置cgroup资源限制
func (c *CgroupManager) Set() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.ContainerID, c.Resource)
	}
	return nil
}

//释放cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
