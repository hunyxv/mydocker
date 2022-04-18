# cgroups

![a.png.jpeg](https://s2.loli.net/2022/04/18/ydNp8QxXEaIRwmM.jpg)

### task
任务（task）。在 cgroups 中，任务就是系统的一个进程；

### subsystem 
subsystem 是一组资源控制模块，可以关联到 cgroup 上，并对 cgroup 中的进程做出相应限制。一般包含下面几项：
- blkio 设置对块设备（比如硬盘）输入输出的访问控制；
- cpu 设置 cgroup 中进程的 CPU 被调度的策略；
- cpuacct 统计 cgroup 中进程的 CPU 占用；
- cpuset 在多核机器上设置 cgroup 中进程可以使用的 CPU ;
- devices 控制 cgroup 中进程对设备的访问；
- freezer 用于挂起和回复进程；
- memory 用于控制 cgroup 中进程的内存占用；
- net_cls 用于将 cgroup 中进程产生的网络包分类 以便于 Linux 中的 tc （traffic controller）可以根据分类区分来自某个 cgroup 的包并作限流或监控；
- net_prio — 这个子系统用来设计网络流量的优先级;
- ns 作用是 使 cgroup 中的进行在新的 Namespace 中 fork 新的进程时，创建出一个新的 cgroup，这个 - cgroup 包含新的 Namespace 前面新的 Namespace 中的进程。



### 控制组 cgroup
一个cgroups包含一组进程，并可以在这个cgroups上增加Linux subsystem的各种参数配置，将一组进程和一组subsystem关联起来

### 层级 hierarchy 
hierarchy 可以把一组 cgroup 串成一个树状结构，这样的树状结构就是 hierarchy， 通过这样的树状结构 cgroups 可以做到继承。

![b.png.jpeg](https://s2.loli.net/2022/04/18/YmAiosaHRufnJQB.jpg)

### subsystem 、cgroup、hierarchy 三者关系
在创建一个新的 hierarchy 时，系统会自动创建一个根 cgroup 并将系统所有进程加入到该 cgroup 中，在这个 hierarchy 中创建的所有 cgroup 都是该根 cgroup 的子节点。
一个 subsystem 只可以分配给一个 hierarchy，而 hierarchy 可以有多个 subsystem
一个进程可以是多个 cgroup 的成员，但这些 cgroup 必须在不同 hierarchy 下
当一个进程创建子进程时，系统会自动将子进程号加入到父进程所属 cgroup 下。

#### /proc 目录
Linux系统上的/proc目录是一种文件系统，即 proc 文件系统。与其它常见的文件系统不同的是，/proc是一种伪文件系统（也即虚拟文件系统），存储的是当前内核运行状态的一系列特殊文件，用户可以通过这些文件查看有关系统硬件及当前正在运行进程的信息，甚至可以通过更改其中某些文件来改变内核的运行状态。 

### 查看有哪些子系统
使用命令 `ll /sys/fs/cgroup` 或 `cat /proc/cgroup`
```
#subsys_name    hierarchy    num_cgroups    enabled
cpuset    8    1    1
cpu    6    60    1
cpuacct    6    60    1
blkio    3    60    1
memory    2    112    1
devices    4    60    1
freezer    11    1    1
net_cls    7    1    1
perf_event    9    1    1
net_prio    7    1    1
hugetlb    5    1    1
pids    10    63    1
rdma    12    1    1
```
各个子系统的说明： 
- cpuset:把任务绑定到特定的cpu
- cpu:    限定cpu的时间份额
- cpuacct: 统计一组task占用cpu资源的报告
- blkio:限制控制对块设备的读写
- memory:  限制内存使用    
- devices: 限制设备文件的创建\限制对设备文件的读写
- freezer: 暂停/恢复cgroup中的task
- net_cls: 用classid标记该cgroup内的task产生的报文
- perf_event: 允许perf监控cgroup的task数据
- net_prio: 设置网络流量的优先级
- hugetlb:  限制huge page 内存页数量
- pids:   限制cgroup中可以创建的进程数
- rdma:  限制RDMA资源(Remote Direct Memory Access，远程直接数据存取)


## 运行

```
go run cmd/cmd.go run -t --memory /bin/bash
```

在启动的进程外查看:

```
cat /sys/fs/cgroup/memory/08e78b67507b4f9e9408981880a47f19/tasks
2500 // bash pid
13321 // 测试用的一个 pid （执行了一个 top）
```

在隔离的进程中查看：

```
cat /sys/fs/cgroup/memory/08e78b67507b4f9e9408981880a47f19/tasks
1
37
42  // 这个是 cat 进程号
```

使用 `ps` 命令查找 2500 进程：

```
ps -ef | grep "2500"
root        2500    2494  0 11:55 pts/0    00:00:00 [bash]
root       13321    2500  0 12:38 pts/0    00:00:00 top
```

由上面可以得出，在 namespace 隔离下，`/bin/bash` 进程号为 1，cgroup 中显示也为 1。