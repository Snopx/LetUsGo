---
title: Kubernetes 容器持久化存储
date: 2022-2-23 16:22:00
tags: [Kubernetes,学习笔记,存储]
category: Kubernetes

---



## 什么是持久化 Volume ？

容器的 Volume，其实就是将一个宿主机的目录与容器内的目录绑定挂在了一起。持久化 Volume 指宿主机上的这个目录，具备 “持久性”。即这个目录的内容，不会因为容器的删除被清理掉，也不会跟当前宿主机绑定，当容器被重启或在其他节点上重建后，它依然能通过挂载这个 Volume 访问到旧内容。

所以大多数情况下，持久化 Volume 的实现，往往依赖一个远程存储服务，比如远程文件存储等。



## 什么是 PV ？

PV 描述的是持久化存储数据卷。这个对象主要是定义一个持久化存储在宿主机的目录。

通常情况下，PV 对象是由运维人员事先创建在 Kubernetes 集群中待用。官方例子如下：

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: task-pv-volume
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
```



## 什么是 PVC？

PVC 是 Pod 希望使用的持久化存储的属性。比如 Volume 存储的大小，可读写权限等。

PVC 通常由开发者创建，或者以 PVC 模板的方式成为 StatefulSet 的一部分，然后由 StatefulSet 控制器负责创建带编号的 PVC。

但是开发者创建的 PVC 真正被容器使用起来，就必须和某个符合条件的 PV 进行绑定。它需要满足两个条件：

1.PV 和 PVC 的 spec 字段，比如 PV 的存储大小必须满足 PVC 的要求。

2.PV 和 PVC 的 storageClassName 字段必须一样。

PVC 和 PV 进行绑定后，Pod 就可以使用 hostPath 等常规类型的 Volume 一样，在自己的 YAML 文件里声明使用这个 PVC：

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    role: web-frontend
spec:
  containers:
  - name: web
    image: nginx
    ports:
      - name: web
        containerPort: 80
    volumeMounts:
        - name: task-pv-volume
          mountPath: "/usr/share/nginx/html"
  volumes:
  - name: task-pv-volume 
    persistentVolumeClaim:
      claimName: task-pv-volume
```

> PVC 可以理解为持久化存储的 “接口” ，它提供了对某种持久化存储的描述，但不提供具体的实现；而这个持久化存储的实现部分则由 PV 负责完成的。



## Volume Controller （专门处理持久化存储的控制器）

### PersistentVolumeController

用于动态配对 PV 和 PVC。它会不断检测每一个 PVC 是不是处于 Bound （已绑定）的状态，如果不是，它就会遍历所有可用的 PV，并尝试绑定。



## 两阶段处理

**第一阶段 （Attach）：Kubelet 需要先调用 API 将它提供的 Persistent Disk 挂载到 Pod 所在的宿主机上。**

**第二阶段（Mount）：格式化整个磁盘设备，然后将它挂载到宿主机指定的挂载点上。**

AttachDetachController 会在第一阶段不断检查每一个 Pod 对应的 PV ，和这个 Pod 所在宿主机之间挂载情况，而决定是否需要对这个 PV 进行 Attach 操作。

AttachDetachController 是一个独立 kubelet 主循环的 Goroutine，用于控制第二阶段操作。



## StorageClass

Kubernetes 为我们提供了一套可以自动创建 PV 的机制，即： Dynamic Provisioning

StorageClass 对象会定义两个部分：

1.PV 的属性。比如存储类型，Volume 的大小

2.PV 需要用到的存储插件。

Kubernetes 能根据用户提交的 PVC ,找到一个对应的 Storage Class ，然后创建出需要的 PV。

Kubernetes 只会将 StorageClass 相同的 PVC 和 PV 绑定起来。



## 学习资料

《深入剖析Kubernetes》