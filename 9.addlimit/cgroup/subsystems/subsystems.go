package subsystems

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	// 返回 subsystem 的名字
	Name() string 
	// 设置某个 cgroup
	Set(path string, res *ResourceConfig) error
	// 将进程添加到某个 cgroup
	Apply(path string, pid int) error
	// 移除某个 cgroup
	Remove() error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubsystem{},
		&MemorySubsystem{},
		&CpuSubsystem{},
	}
)