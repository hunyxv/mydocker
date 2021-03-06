# rootfs
根文件系统（rootfs）首先是内核启动时所mount的第一个文件系统，内核代码映像文件保存在根文件系统中，而系统引导启动程序会在根文件系统挂载之后从中把一些基本的初始化脚本和服务等加载到内存中去运行。 -- 百度百科

# chroot

`chroot` 系统调用能够将当前进程的根目录更改为一个新的目录，新的根目录还将被当前进程的所有子进程所继承。在不需要 `mount namespace` 的情况下实现切换容器内进程根目录的效果。


# pivot_root

`pivot_root` 系统调用，更改调用进程的 mount 命名空间中的根文件系统。 更准确地说，它将根挂载移动到目录 put_old 并使 new_root 新的根挂载。 调用进程必须在拥有调用方的 mount 命名空间的用户命名空间中具有 CAP_SYS_ADMIN 功能。

将同一 mount 命名空间中每个进程或线程的根目录和当前工作目录更改为 new_root（如果它们指向旧的根目录）。另一方面，`pivot_root()` 不会更改调用方的当前工作目录（除非它位于旧的根目录上），因此它应该后跟 chdir（“/”） 调用。

# 启动
1. 准备一基础镜像 `ubuntu.tar`:

``` shell
docker pull ubuntu:16.04

docker run -d ubuntu:16.04 tail -f /dev/null

docker export -o ubuntu.tar a7b8788ff3dc
```

使用 ubuntu.tar 作为 rootfs

```
go run cmd/cmd.go run -t -i ubuntu.tar /bin/bash
```