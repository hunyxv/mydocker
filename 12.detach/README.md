# 后台运行 container 中的程序

使用 `-d` or `--detach` 参数，使容器中的进程放在后台执行。考虑到启动参数中如果有 `--rm` 容器进程结束后需要执行一些清理操作。所以实现 detash 参数功能的逻辑很简单，就是主进程 fock 出一个子进程执行启动容器的操作，然后主进程退出，那么子进程就变成了孤儿进程，进程号为 1 的进程会接收这个孤儿进程。这个孤儿进程会 fock 一个子进程执行容器中的程序，那么这个孤儿进程就是容器进程的守护进程，守护进程等待容器进程结束后，执行清理等操作，然后结束进程。

启动命令(示例)： `go run cmd/cmd.go run  -i ./ubuntu.tar --rm -d -- /bin/bash -c "sleep 3; echo 'hello world'; sleep 1000"`

> 其中 `-- /bin/bash -c "xxx"` `--` 符号表示其后面的字符串都作为参数传递

标准输出被收集在日志文件中：`/var/lib/mydocker/[container id]/log.json`。
守护进程信息在 `/var/lib/mydocker/containers/[container id].json` 文件中，格式：

```json
{
    "pid": 207345,
    "id": "218ea2e563644bfba77beda183ea28cc",
    "name": "218ea2e563644bfba77beda183ea28cc",
    "command": "-- /bin/bash -c sleep 3; echo 'hello world'; sleep 1000",
    "createTime": "2022-04-22 18:33:58.217574753 +0800 CST m=+0.338664010",
    "status": "",
    "volume": [],
    "portmapping": null
}
```

使用 `ps -ef | grep 207345` 查找此进程（207345）的父进程：

```
$ ps -ef | grep 207345
root      207345  207340  0 18:33 ?        00:00:00 /bin/bash -c sleep 3; echo 'hello world'; sleep 1000
root      207373  207345  0 18:34 ?        00:00:00 sleep 1000
```

父进程ID为 207340，查找 207340 的父进程ID，为 1:

```
root      207340       1  0 18:33 ?        00:00:00 /proc/self/exe run --image ./ubuntu.tar --rm -- /bin/bash -c sleep 3; echo 'hello world'; sleep 1000
root      207345  207340  0 18:33 ?        00:00:00 /bin/bash -c sleep 3; echo 'hello world'; sleep 1000
```