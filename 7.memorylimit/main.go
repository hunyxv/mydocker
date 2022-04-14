package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

/*
	/proc/self/exe 是指当前程序

	go run main.go // os.Args[0] 是 /tmp/go-build1788250820/b001/exe/main 

	然后通过 exec.Command 启动一个进程执行 `/proc/self/exe`,也就是当前程序自己，而 os.Args[0]
	就是 /proc/self/exe
*/

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func main() {
	fmt.Println(os.Args[0])
	if os.Args[0] == "/proc/self/exe" {
		fmt.Printf("current pid %d", syscall.Getpid())
		fmt.Println()
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")  // 执行当前程序  os.Args[0] 就是 /proc/self/exe
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("pid %v\n", cmd.Process.Pid)

	// 在系统默认创建挂载了 memory subsystem 的 Hierarchy 上创建 cgroup
	os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "test_memory_limit"), 0755)
	// 将进程加入到 cgroup
	ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "test_memory_limit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	// 限制 cgroup 进程的使用
	ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "test_memory_limit", "memory.limit_in_bytes"), []byte("100m"), 0644)
	cmd.Process.Wait()
}

