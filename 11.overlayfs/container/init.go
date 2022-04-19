package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
)

func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command %s", command)
	err := setUpMount()
	if err != nil {
		return err
	}

	cmdpath, err := exec.LookPath(command)
	if err != nil {
		return err
	}

	if err = syscall.Exec(cmdpath, args, []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"}); err != nil {
		logrus.WithError(err).Error("exec err")
		return err
	}
	return nil
}

// 设置挂载点
func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.WithError(err).Error("os.Getwd err")
		return err
	}
	logrus.Infof("current location is %s", pwd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", pwd+"/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		logrus.WithError(err).Error("mount '/proc' fail")
		return err
	}

	if err := pivotRootFS(pwd); err != nil {
		logrus.WithError(err).Error("pivot root mount fail")
		return err
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		logrus.WithError(err).Error("mount '/dev' fail")
		return err
	}
	return nil
}

// 系统调用 -- pivot_root()
// 	将同一 mount 命名空间中每个进程或线程的根目录和当前工作目录更改为 new_root
func pivotRootFS(newroot string) error {
	if err := syscall.Mount(newroot, newroot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	// 创建 rootfs/.pivot_root 存储 old_root
	putold := filepath.Join(newroot, ".pivot_root")
	if err := os.Mkdir(putold, 0700); err != nil {
		return err
	}

	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return fmt.Errorf("saycall.PivotRoot %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	// 将 old_root 从 ’/.pivot_root‘ 上卸载，并删除此目录
	putold = "/.pivot_root"
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	return os.Remove(putold)
}
