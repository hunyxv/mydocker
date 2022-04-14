# 一个简单的容器

## linux 系统 /proc 目录介绍

在 GUN/Linux 操作系统中的/proc是一个位于内存中的伪文件系统(或者叫做虚拟文件系统)。该目录下保存的不是真正的文件和目录，而是一些"运行时"的信息，例如系统内存、磁盘IO、设备挂载信息和硬件配置信息等。proc目录是一个控制中心，用户可以通过更改其中某些文件来改变内核的运行状态，proc目录也是内核提供给一个的查询中心，可以通过这些文件查看有关系统硬件及当前正在运行进程的信息。在 Linux 系统中，许多工具的数据来源正是 proc 目录中的内容，例如：lsmod命令就是 `cat /proc/modules` 命令的别名，lspci 命令是 `cat /proc/pci` 命令的别名

简单一点来讲，/proc 目录就是保存在系统中的信息，其包含许多以数字命名的子目录，这些数字代表着当前系统正在运行进程的进程号，里面包含对应进程相关的多个信息文件

几个文件/目录：
- /proc/pid
    每一个 /proc/pid 目录中还存在一系列目录和文件, 这些文件和目录记录的都是关于 pid 对应进程的信息. 例如在 /proc/pid 的目录下存在一个 task 目录, 在 task 目录下又存在 task/tid 这样的目录, 这个目录就是包含此进程中的每个线程的信息, 其中的 tid 是内核线程的 tid, 通过 GETDENTS 遍历 /proc 就能够看到所有的 /proc/pid 的目录, 当然通过 ls -al /proc 的方式也可以看到所有的信息
- /proc/self
    这是一个 link, 当进程访问此 link 时, 就会访问这个进程本身的 /proc/pid 目录。
- /proc/self/exe
    这个就是当前运行的程序本体(软连接到程序)

...

## simple container 

```
[root@xxxxx 8.simplecontainer]# go run cmd/cmd.go -h
The purpose of this project is to learn how docker works and how to write a docker by ourselves.
        Enjoy it, just for fun.

Usage:
  mydocker [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        I nit container process run user's process in container . Do not call it outside
  run         Create a container with namespace and cgroups limit mydocker run -it [command]
```

run 指令会创建一个在隔离了宿主机命名空间的进程（pid=1），然后执行自动执行 init 指令。
init 指令首先在隔离的环境中挂载 proc 文件系统，然后通过系统调用 exec 执行bash命令，exec 并不会创建新的进程，也就是说新执行的进程替换了上面 run 启动的进程，pid = 1.

> exec command in Linux is used to execute a command from the bash itself. This command does not create a new process it just replaces the bash with the command to be executed. If the exec command is successful, it does not return to the calling process.


比如执行 `[root@xxxxx 8.simplecontainer]# go run cmd/cmd.go run -t /bin/bash` 然后执行 `top` 查看进程情况：

```
[root@xxxx 8.simplecontainer]# go run cmd/cmd.go run -t /bin/bash
INFO[0000] command /bin/bash 
[root@xxxx 8.simplecontainer]# top
top - 17:52:52 up 37 days, 22:11,  2 users,  load average: 0.71, 0.61, 0.51
Tasks:   2 total,   1 running,   1 sleeping,   0 stopped,   0 zombie
%Cpu(s):  8.7 us,  4.3 sy,  0.0 ni, 86.0 id,  0.0 wa,  0.7 hi,  0.3 si,  0.0 st
MiB Mem :   1735.8 total,    239.8 free,    645.3 used,    850.7 buff/cache
MiB Swap:    500.0 total,    163.5 free,    336.4 used.    927.2 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND                                                            
      1 root      20   0   27572   5240   3336 S   0.0   0.3   0:00.01 bash                                                               
     21 root      20   0   65304   4384   3772 R   0.0   0.2   0:00.00 top 
```

可以看到只用两个进程（bash、top），top 命令就是读取的 `/proc` 下的信息。