package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pborman/uuid"
)



type ContainerInfo struct {
	Pid         int      `json:"pid"`         //容器的init进程在宿主机上的 PID
	Id          string   `json:"id"`          //容器Id
	Name        string   `json:"name"`        //容器名
	Command     string   `json:"command"`     //容器内init运行命令
	CreatedTime string   `json:"createTime"`  //创建时间
	Status      string   `json:"status"`      //容器的状态
	Volume      []string `json:"volume"`      //容器的数据卷
	PortMapping []string `json:"portmapping"` //端口映射
}

type ParentProcess struct {
	cmd  *exec.Cmd
	logf *os.File

	autoRemove  bool
	ttl         bool
	workSpace   *WorkSpace
	containerid string
	volumes     []string
}

func NewParentProcess(ttl bool, command []string, image string, rm bool, volumes []string) (*ParentProcess, error) {
	cid := strings.ReplaceAll(uuid.NewRandom().String(), "-", "")

	workSpace, err := NewWorkSpace(image, cid)
	if err != nil {
		return nil, err
	}
	return &ParentProcess{
		cmd: newcommand(command),

		containerid: cid,
		workSpace:   workSpace,
		autoRemove:  rm,
		ttl:         ttl,
		volumes:     volumes,
	}, nil
}

func (pproc *ParentProcess) ContainerID() string {
	return pproc.containerid
}

func (pproc *ParentProcess) Start() error {
	if err := pproc.workSpace.CreateWorkSpace(); err != nil {
		return err
	}

	cmd := pproc.cmd
	if pproc.ttl {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// TODO
		logPath := pproc.workSpace.containerFS + "/log.json"
		logf, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			return fmt.Errorf("create log file fail, %v", err)
		}

		pproc.logf = logf
		SetLogToFile(logf)
		cmd.Stdout = &StdoutRedirect{}
		cmd.Stderr = &StderrRedirect{}
	}
	cmd.Dir = pproc.workSpace.RootPath()
	for _, v := range pproc.volumes {
		dirs := strings.Split(v, ":")
		if len(dirs) == 2 {
			err := pproc.workSpace.mountVolume(dirs[0], dirs[1])
			if err != nil {
				return err
			}
		}
	}
	return cmd.Start()
}

func (pproc *ParentProcess) Wait() error {
	return pproc.cmd.Wait()
}

func (pproc *ParentProcess) PID() (int, error) {
	if pproc.cmd.Process == nil {
		return 0, fmt.Errorf("cmd not start")
	}
	return pproc.cmd.Process.Pid, nil
}

func (pproc *ParentProcess) Release() error {
	if !pproc.ttl && pproc.logf != nil {
		//logPath := pproc.logf.Name()
		pproc.logf.Close()
	}
	if pproc.autoRemove {
		err := pproc.workSpace.Remove()
		if err != nil {
			return err
		}
	}

	return nil
}

func newcommand(command []string) *exec.Cmd {
	args := append([]string{"init",}, command...)
	cmd := exec.Command("/proc/self/exe", args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0, // 0 表示 root，只用 root 才能执行 init 中的 mount proc 操作
				HostID:      os.Geteuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getegid(),
				Size:        1,
			},
		},
	}
	return cmd
}
