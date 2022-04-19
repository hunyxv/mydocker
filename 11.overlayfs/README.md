# OverlayFS 文件系统

OverlayFS 是类似 AUFS 的现代联合文件系统（union filesystem），但是速度更快，实现更简单。针对 OverlayFS 提供了两个存储驱动：最初的 `overlay`，以及更新更稳定的 `overlay2`。

OverlayFS 层（layers） 在单个 Linux 主机上分为两个目录，并且将它们呈现为单个目录。这些目录统称为层（layers），统一过程称为联合挂载（union mount）。OverlayFS 把下层目录称为 lowerdir，上层目录称为 upperdir，统一视图通过称为 merged 自身目录暴露。

overlay 驱动仅适用单个 lower OverlayFS 层，因此需要通过硬链接来实现多层镜像，overlay2 驱动原生支持 128 个 lower OverlayFS 层。这个功能为与层相关的命令如 docker build 和 docker commit 提供了更好的性能，并且在后备文件系统上消耗更少的 inode。



## docker 镜像和容器层
docker pull 下一个 ubuntu 系统镜像后，可以看到 image 有 4 层：

```bash
root@hunyxv-VirtualBox:~/docker# docker pull ubuntu:16.04
16.04: Pulling from library/ubuntu
58690f9b18fc: Pull complete
b51569e7c507: Pull complete
da8ef40b9eca: Pull complete
fb15d46c38dc: Pull complete
Digest: sha256:0f71fa8d4d2d4292c3c617fda2b36f6dabe5c8b6e34c3dc5b0d17d4e704bd39c
Status: Downloaded newer image for ubuntu:16.04
docker.io/library/ubuntu:16.04
```
对应 `/var/lib/docker/overlay2`  目录下也有 4 个文件夹和一个名为 l 的文件夹，其中 l 目录包含缩短的层标识符作为软链接，这些标识符用于避免 `mount` 命令参数页面大小限制：

```bash
root@hunyxv-VirtualBox:~/docker# ls /var/lib/docker/overlay2/
5bc64c1f516809249da96df876bdd17661b827b09651b27b1994fd784b68b8b2
72278702e7e93864288b124d7758b79e27f03275da4613f9878c42bc613b682d
7f4afce98a4f6904f9038995b08ce536e4e1ae048761c9961657e06cd90504ea
c741490dd4f947c6c8cee95142b091f483d6f9fec257df17deb3a60a498962ae
l

root@hunyxv-VirtualBox:~/docker# ls -l /var/lib/docker/overlay2/l
total 16
lrwxrwxrwx 1 root root 72  4月 13 12:52 GMBJQEJAS5KOMOQNHVJ6FZWUUY -> ../72278702e7e93864288b124d7758b79e27f03275da4613f9878c42bc613b682d/diff
lrwxrwxrwx 1 root root 72  4月 13 12:52 LFRZ2OW6KV55T3YLHLMYIHQMVQ -> ../5bc64c1f516809249da96df876bdd17661b827b09651b27b1994fd784b68b8b2/diff
lrwxrwxrwx 1 root root 72  4月 13 12:52 VGINZP4WW66IDQY3ZJIUS2JS7G -> ../7f4afce98a4f6904f9038995b08ce536e4e1ae048761c9961657e06cd90504ea/diff
lrwxrwxrwx 1 root root 72  4月 13 12:52 XGACNPNPHNVF5PTWXGDZMXLCPZ -> ../c741490dd4f947c6c8cee95142b091f483d6f9fec257df17deb3a60a498962ae/diff
```
然后我们使用这个继承镜像构建一个新的镜像（添加一次层）:

```bash
# Dockerfile:
FROM ubuntu:16.04

RUN echo "Hello World" > /tmp/newfile

# build 镜像
root@hunyxv-VirtualBox:~/docker#  docker build -t changed-ubuntu .
Sending build context to Docker daemon  5.632kB
Step 1/2 : FROM ubuntu:16.04
 ---> b6f507652425
Step 2/2 : RUN echo "Hello World" > /tmp/newfile
 ---> Running in c6294cce6a40
Removing intermediate container c6294cce6a40
 ---> 5c2ac708085c
Successfully built 5c2ac708085c
Successfully tagged changed-ubuntu:latest

# 启动
docker run -d changed-ubuntu sleep 360000000
```
再查看 /var/lib/docker/overlay2 下的目录，可以看到多了几个：

```bash
root@hunyxv-VirtualBox:~/docker# ls /var/lib/docker/overlay2/
567a12b2accb443b9b0a8624868bd11a4be3afc020e49a60e5ec43c899385cc6   // 这个是容器启动后出现的
567a12b2accb443b9b0a8624868bd11a4be3afc020e49a60e5ec43c899385cc6-init // 容器删除后，即消失
5bc64c1f516809249da96df876bdd17661b827b09651b27b1994fd784b68b8b2
72278702e7e93864288b124d7758b79e27f03275da4613f9878c42bc613b682d
7f4afce98a4f6904f9038995b08ce536e4e1ae048761c9961657e06cd90504ea
819f092c08ae13cd6ef3c8d0d7739635f1086bf6ff1574a6b22cfa838b07912f  // 这个就是上面新加的一层
c741490dd4f947c6c8cee95142b091f483d6f9fec257df17deb3a60a498962ae
l
```


我们查看新加的一次，每一层目录中都包含一个文件 `lower`  ，内容为此层的父级，以及包含这层镜像内容的名为 `diff` 的目录，在 `diff` 目录中可以看到上面的改动（`/tmp/newfile`）。它包含一个 `merged` 目录，包括父层以及自身的统一内容，以及 OverlayFS 自身使用的 `work` 目录。

```bash
root@hunyxv-VirtualBox:~/docker# ls /var/lib/docker/overlay2/819f092c08ae13cd6ef3c8d0d7739635f1086bf6ff1574a6b22cfa838b07912f/
committed  diff  link  lower  work

root@hunyxv-VirtualBox:~/docker# cat /var/lib/docker/overlay2/819f092c08ae13cd6ef3c8d0d7739635f1086bf6ff1574a6b22cfa838b07912f/lower
l/VGINZP4WW66IDQY3ZJIUS2JS7G:l/LFRZ2OW6KV55T3YLHLMYIHQMVQ:l/XGACNPNPHNVF5PTWXGDZMXLCPZ:l/GMBJQEJAS5KOMOQNHVJ6FZWUUY

root@hunyxv-VirtualBox:~/docker# ls /var/lib/docker/overlay2/819f092c08ae13cd6ef3c8d0d7739635f1086bf6ff1574a6b22cfa838b07912f/diff/
tmp
```
通过 `mount` 命令查看 Docker 使用 `overlay2` 存储驱动的挂载情况：

```bash
root@hunyxv-VirtualBox:~/docker# mount | grep "overlay"
overlay on /var/lib/docker/overlay2/567a12b2accb443b9b0a8624868bd11a4be3afc020e49a60e5ec43c899385cc6/merged type overlay (rw,relatime,lowerdir=/var/lib/docker/overlay2/l/V367XM7CTSIFKCYS6H3WEDBW3M:/var/lib/docker/overlay2/l/4LGRXWTBZDJZWFZHCZ4HRYOAMO:/var/lib/docker/overlay2/l/VGINZP4WW66IDQY3ZJIUS2JS7G:/var/lib/docker/overlay2/l/LFRZ2OW6KV55T3YLHLMYIHQMVQ:/var/lib/docker/overlay2/l/XGACNPNPHNVF5PTWXGDZMXLCPZ:/var/lib/docker/overlay2/l/GMBJQEJAS5KOMOQNHVJ6FZWUUY,upperdir=/var/lib/docker/overlay2/567a12b2accb443b9b0a8624868bd11a4be3afc020e49a60e5ec43c899385cc6/diff,workdir=/var/lib/docker/overlay2/567a12b2accb443b9b0a8624868bd11a4be3afc020e49a60e5ec43c899385cc6/work)
```
`rw` 选项显示 `overlay` 是读写方式挂载的。



## overlay 驱动是如何工作的
OverlayFS 层（layers） 在单个 Linux 主机上分为两个目录，并且将它们呈现为单个目录。这些目录统称为层（layers），统一过程称为联合挂载（union mount）。OverlayFS 把下层目录称为 lowerdir，上层目录称为 upperdir，统一视图通过称为 merged 自身目录暴露。

下图展示了一个 Docker 镜像和一个 Docker 容器如何分层。镜像层术语 lowerdir，容器层术语 upperdir。统一视图通过名为 merged 的目录暴露。

![L7IDH8jzaxT0zZ-icvFnaZWem1joMC4oyGEW5eImMrM.png](https://s2.loli.net/2022/04/19/nDbVioTxzUtjw1f.png)

在镜像层和容器层都包含相同文件时，则容器层为主，并且掩盖镜像层同一个文件的存在。

overlay 驱动仅适用于两层，这意味着多层镜像不能实现多个 OverlayFS 层。取而代之，每个镜像层都在 /var/lib/docker/overlay 下实现自己的目录。通过硬链接引用与底层共享数据的方式来节省空间。从 Docker 1.10 开始，镜像层 IDs 不再对应于 /var/lib/docker 中的目录名。

为了创建一个容器，overlay 驱动组合顶层的目录以及容器的新目录。镜像的顶层是叠加层中的 lowerdir，并且是只读挂载的。容器的新目录是 upperdir 并且是可写的。



### 手动挂载一个 overlay 目录
Upper -> Lower 的次序一次堆叠起来，最终用户看到的就是 MergedDir。从上到下，用户只能看到最上层的文件，而无法越过它。这就是实现了一种类似文件夹合并的策略。用一句话理解就是 上层文件优先。

它是零拷贝的，它的速度非常快，只需要执行一次 mount 操作就能实现。

另外，你可能注意到了，在用户看到的最终文件系统层下，有一个 UpperDir，这是你在此文件系统中唯一可以操作变动的层，你的所有 增加、修改、删除 操作都是在这一层上完成的，不会有任何记录留在 LowerDir 中，这可以保证你的操作不会污染到原文件夹，即使你看起来是删除的是来自于 LowerDir 的文件。



首先初始创建这样的目录和文件：

```bash
overlay
├── lower1
│   └── lower1.txt
├── lower2
│   └── lower3.txt
├── lower3
│   └── lower3.txt
├── merged
├── upper
│   └── upper.txt
└── work
```
* lower1、lower2、upper 三个文件夹，这个三个文件夹是用来合并的
* 最终的文件操作都会存储在 upper 中
* 一个空的 worker 文件夹，这文件夹不能有任何内容
* 最后需要一个 merged 文件夹，用来作为给用户呈现的最终文件夹（挂载点，卸载后本文件夹中无数据）



执行挂载命令：

```bash
mount -t overlay -o lowerdir=lower1:lower2:lower3,upperdir=upper,workdir=work overlay merged

# 挂载后 merged 效果如上图所示：
root@hunyxv-VirtualBox:/tmp/overlay# ls merged/
lower1.txt  lower3.txt  upper.txt
```
`ower1:lower2:lower3`是指定lower层的目录，越前面越优先，`lower1 > lower2 > lower3`。这里需要注意，lower层可以指定非常多，但是在 overlay1 版本中，最多支持 128 层，在 overlayfs2 中添加了更多层的支持。

增加文件：

```bash
root@hunyxv-VirtualBox:/tmp/overlay# touch merged/merged.txt
root@hunyxv-VirtualBox:/tmp/overlay# tree .
.
├── lower1
│   └── lower1.txt
├── lower2
│   └── lower3.txt
├── lower3
│   └── lower3.txt
├── merged
│   ├── lower1.txt
│   ├── lower3.txt
│   ├── merged.txt  // 新建文件
│   └── upper.txt
├── upper
│   ├── merged.txt
│   └── upper.txt // 可以看到，增加文件是直接添加到了 upper 文件夹中
└── work
    └── work
```


删除文件:

```bash
root@hunyxv-VirtualBox:/tmp/overlay# rm merged/lower1.txt
root@hunyxv-VirtualBox:/tmp/overlay# tree .
.
├── lower1            // 而 lower1 文件夹中的文件并没有真正的被删除
│   └── lower1.txt
├── lower2
│   └── lower2.txt
├── lower3
│   └── lower3.txt
├── merged            // lower1.txt 文件已经不存在了
│   ├── lower2.txt
│   ├── lower3.txt
│   ├── merged.txt
│   └── upper.txt
├── upper
│   ├── lower1.txt  // c--------- 2 root root 0, 0  4月 13 15:36 lower1.txt 变成了一个特殊类型文件
│   ├── merged.txt
│   └── upper.txt
└── work
    └── work
        └── #76
```
upper 中的 多了一个特殊类型文件（字符设备）：

```bash
c--------- 2 root root 0, 0  4月 13 15:36 lower1.txt
```


修改文件

```bash
root@hunyxv-VirtualBox:/tmp/overlay# echo "testttttt" > merged/lower2.txt
root@hunyxv-VirtualBox:/tmp/overlay# tree .
.
├── lower1
│   └── lower1.txt
├── lower2
│   └── lower2.txt
├── lower3
│   └── lower3.txt
├── merged
│   ├── lower2.txt
│   ├── lower3.txt
│   ├── merged.txt
│   └── upper.txt
├── upper
│   ├── lower1.txt
│   ├── lower2.txt       // 多了一个文件
│   ├── merged.txt
│   └── upper.txt
└── work
    └── work
        └── #7c
```
当文件修改后，我们查看上层的 upper 文件夹，会发现，多了一个 lower2.txt。而底层的 lower2.txt 仍然没有内容变化。依此可以推断出，其实修改文件是将下层的文件复制到上层来之后再进行修改的。



## 容器是如何通过 overlay 或 overlay2 读写的
**读取文件**

考虑三个容器通过 overlay 打开文件读取的场景。

* 容器层中不存在该文件：如果容器打开读取一个并不存在容器层（upperdir），则从镜像层（lowerdir）读取该文件。这会导致很少的性能开销。
* 文件仅存在于容器层：如果容器打开读取一个存在于容器层（upperdir）但不存在于镜像层（lowerdir）的文件，则直接从容器层中读取该文件。
* 该文件同时存在于容器层和镜像层：如果容器打开读取一个同时存在于容器层（upperdir）和镜像层（lowerdir）的文件，则容器层（upperdir）会覆盖镜像层（lowerdir） 相同名字的文件。



**修改文件或目录**

同样分三个场景来介绍修改：

* 第一次写入文件：容器第一次写入现有文件时，这个文件还不存在于容器层（upperdir）。overlay/overlay2 驱动程序从镜像层（lowerdir）执行一个 copy\_up 操作到容器层（upperdir）。然后，容器将更改写入容器层中的文件的新副本。但是，OverlayFS 工作在文件级别而不是块级别，意味着所有 OverlayFS copy\_up 操作都会复制整个文件，即使文件非常大，并且只修改了其中的一小部分。这就对容器写入性能产生显著的影响。不过，有两件事值得注意：
   * copy\_up 操作仅在第一次写入文件时发生，对同一文件的后续写入操作只对已复制到容器的文件副本进行操作。
   * OverlayFS 仅适用于两层，意味着它性能应该是优于 AUFS 的，当搜索多个层的镜像文件时，AUFS 会出现明显的延迟。这个优势适用于 overlay 和 overlay2，overlayfs2 在初始读取时的性能略低于 overlayfs，因为它会查看更多层级，但是它会缓存结果。
* 删除文件或者目录：在容器中删除文件时，会在容器层（upperdir）中创建一个 whiteout 文件。镜像层（lowerdir）中文件的版本并不会被删除（因为 lowerdir 是只读的）。但是，whiteout 文件会阻止其在容器中可用。当在容器中删除目录时，会在容器层（upperdir）中创建一个 opaque 目录。它的工作机制同 whiteout，并且有效地防止目录被访问，即使它仍然存在于镜像层（lowerdir）。
* 重命名目录：仅当源路径和目标路径都在顶层时，才允许目录调用 rename(2)，否则会返回 EXDEV error（"cross-device link not permitted"）。


# 使用 overlayfs

1. 准备一基础镜像 `ubuntu.tar`:

``` shell
docker pull ubuntu:16.04

docker run -d ubuntu:16.04 tail -f /dev/null

docker export -o ubuntu.tar a7b8788ff3dc
```

2. 在当前路径执行 `go run cmd/cmd.go run -t -i ./ubuntu.tar --rm /bin/bash`