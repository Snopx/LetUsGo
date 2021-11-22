---
title: Kubernetes Pod
date: 2021-11-18 22:35:00
tags: [Kubernetes,学习笔记]
category: Kubernetes
---

## 什么是 Pod?

Pod 是 Kubernetes 调度的最小单位，容器的集合，一组紧密相关的容器放在一个Pod中，同一个Pod 中的所有容器共享IP地址和 Port空间，他们在一个 network namespace 中。同一Pod中的容器始终被一起调度。



## **为什么需要Pod**

在实际部署的应用中，往往会存在类似 “进程和进程组”的关系,**因为这些应用之间是紧密联系的，它们必须部署在同一台机器上。**

Pod 是 Kubernetes 里原子调度单位，这意味着它的调度器是统一按照 Pod 而非容器的资源需求进行计算的。（而是Pod 需要的资源）

容器见的紧密协作指的是：互相之间会发生直接的文件交换，使用 [localhost](http://localhost) 或者 Socket 文件进行本地通信，会发生非常频繁的调用等

PS:并不是所有有关系的容器都应该属于同一个 Pod ,比如 Mysql。

**Pod 只是一个逻辑概念.**

Kubernetes 真正处理的，还是宿主机操作系统上 Linux 容器的 Namespace 和 Cgroups,而并不存在一个所谓的 Pod 的边界或者隔离环境。

**Pod,其实是一组共享了某些资源的容器。这些容器共享同一个 NetWork Namespace,并且可以声明共享同一个 Volume。**



> 如果要为 Kubernetes 开发一个网络插件时，应该重点考虑的是如何配置这个 Pod 的 Network Namespace，而不是每一个用户容器如何使用你的网络配置，这是没有意义的。 这意味着网络插件完全不用关心用户容器的启动与否，只需要关注如何配置 Pod，也就是 Infra 容器的 Network Namespace

Pod 这种“超亲密关系”容器的设计思想，实际上就是希望，当用户想在一个容器里跑多个功能并不相关的应用时，应该优先考虑它们是不是更应该被描述成一个 Pod 里的多个容器。



## Pod 中重要字段的含义和用法

NodeSelector ：将 Pod 与 Node 进行绑定，意味着这个 Pod 永远只能运行在携带了 某个标签的节点上;否则，它将调度失败。

NodeName : 一旦 Pod 的这个字段被赋值，Kubernetes 项目就会被认为这个 Pod 已经过了调度，调度的结果就是赋值的节点名字。这个字段一般由调度器负责设置。

HostAliases ：定义 Pod 的 hosts 文件（比如/etc/hosts）里的内容。如果要设置 hosts 文件里的内容，一定要通过这种方法。否则，如果直接修改了 hosts 文件的话，在 Pod 被删除重建之后，kubelet 会自动覆盖掉被修改的内容。

```yaml
apiVersion: v1
kind: Pod
...
spec:
  hostAliases:
  - ip: "10.1.2.3"
    hostnames:
    - "foo.remote"
    - "bar.remote"
...

cat /etc/hosts
# Kubernetes-managed hosts file.
127.0.0.1 localhost
...
10.244.135.10 hostaliases-pod
10.1.2.3 foo.remote
10.1.2.3 bar.remote
```

Init Containers ：优先所有的 Containers ，并严格按照定义的顺序执行。



**Pod 设计目的**

Pod 的设置就是为了让它里面的容器尽可能多共享 Linux Namespace，仅仅保留必要的隔离和限制能力。

在某个 Pod 中，只要定义了共享主机的 NetWork ，IPC 和 PID Namespace，这就意味着，这个 Pod 里面的所有容器，会直接使用宿主机的网络，直接与宿主机进行 IPC 通信 ，看到宿主机里面正在运行的所有进程。

Volume

在 Kubernetes 中，有几种特殊的 Volume，它们存在的意义不是为了存放容器里的数据，也不是用来进行容器和宿主机之间的数据交换。这些特殊 Volume 的作用，是为容器提供预先定义好的数据。

目前为止 , Kubernetes 支持的 Projected Volume 一共有四种：

Secret：帮助把 Pod 想要访问的加密数据，存放到 Etcd 中。可以通过 Pod 的容器里挂载 Volume 的方式，访问到这些 Secret 里保存的信息。

> Secret 对象要求这些数据必须是经过 Base64 转码的，避免出现明文密码的安全隐患。 在真正的生产环境中，你需要在 Kubernetes 中开启 Secret 的加密插件。

像这样通过挂载方式进入到容器里的 Secret ， 一旦其对应的 Etcd 里的数据被更新，这些 Volume 里的文件内容，同样也会被更新。

**需要注意的是：这个更新可能会有一定的延时。所以在编写应用程序时，在发起数据库连接的代码处下写好重试和超时的逻辑。**

ConfigMap：它保存的是不需要加密的，应用所需要的配置信息。用法几乎与 Secret 完全相同。

Downward API：它能让 Pod 里的容器能够直接获取到这个 Pod API 对象本身的信息。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-downwardapi-volume
  labels:
    zone: us-est-coast
    cluster: test-cluster1
    rack: rack-22
spec:
  containers:
    - name: client-container
      image: k8s.gcr.io/busybox
      command: ["sh", "-c"]
      args:
      - while true; do
          if [[ -e /etc/podinfo/labels ]]; then
            echo -en '\\n\\n'; cat /etc/podinfo/labels; fi;
          sleep 5;
        done;
      volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
          readOnly: false
  volumes:
    - name: podinfo
      projected:
        sources:
        - downwardAPI:
            items:
              - path: "labels"
                fieldRef:
                  fieldPath: metadata.labels
```

volumeMounts 中的 mountPath 指定了目录，downwardAPI 的 Items 下 path 参数指定了目录文件，所以会被 Kubernetes 自动挂载成为容器里的 /etc/podinfo/labels 文件。

不过，需要注意的是，Downward API 能够获取到的信息，一定是 Pod 里的容器进程启动之前就能够确定下来的信息。而如果你想要获取 Pod 容器运行后才会出现的信息，比如，容器进程的 PID，那就肯定不能使用 Downward API 了，而应该考虑在 Pod 里定义一个 sidecar 容器。

所以，一般情况下，建议你使用 Volume 文件的方式获取配置信息。

Service Account 对象的作用，就是 Kubernetes 系统内置的一种 “服务账户”, 它是 Kubernetes 进行权限分配的对象。限制各种各种 Service Account 对 Kubernetes API 的操作权限。

像这样的 Service Account 的授权信息和文件，实际上保存在它所绑定的一个特殊的 Secret 对象里的。这个特殊的 Secret 对象，就叫作 ServiceAccountToken。任何运行在 Kubernetes 集群上的应用，都必须使用这个 ServiceAccountToken 里保存的授权信息，也就是 Token，才可以合法地访问 API Server。

**这种把 Kubernetes 客户端以容器的方式运行在集群里，然后使用 default Service Account 自动授权的方式，被称作“InClusterConfig”，也是最推荐的进行 Kubernetes API 编程的授权方式。**



## Pod 容器健康检查和恢复机制

在 Kubernetes 中，你可以为 Pod 里的容器定义一个健康检查“探针”（Probe）。这样，kubelet 就会根据这个 Probe 的返回值决定这个容器的状态，而不是直接以容器镜像是否运行（来自 Docker 返回的信息）作为依据。这种机制，是生产环境中保证应用健康存活的重要手段。

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    test: liveness
  name: test-liveness-exec
spec:
  containers:
  - name: liveness
    image: busybox
    args:
    - /bin/sh
    - -c
    - touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600
    livenessProbe:
      exec:
        command:
        - cat
        - /tmp/healthy
      initialDelaySeconds: 5
      periodSeconds: 5
```

Pod 中的 livenessProbe（健康检测）定义执行一行命令，来查看 /tmp/helthy 文件。在容器启动 5 s 后开始执行（initialDelaySeconds: 5），每 5 s 执行一次（periodSeconds: 5）。这个健康检查，以查看该文件是否存在来判断 Pod 是否正常。

如果健康检查探查到 /tmp/healthy 已经不存在了，那么 Kubernetes 会**重启，但实际确实重新创建了容器。**

这个功能就是 Kubernetes 里的 Pod 恢复机制，也叫 restartPolicy。它是 Pod 的 Spec 部分的一个标准字段（pod.spec.restartPolicy），默认值是 Always，即：任何时候这个容器发生了异常，它一定会被重新创建。

但一定要强调的是，Pod 的恢复过程，永远都是发生在当前节点上，而不会跑到别的节点上去。事实上，一旦一个 Pod 与一个节点（Node）绑定，除非这个绑定发生了变化（pod.spec.node 字段被修改），否则它永远都不会离开这个节点。这也就意味着，如果这个宿主机宕机了，这个 Pod 也不会主动迁移到其他节点上去。





## 学习资料

《深入剖析Kubernetes》