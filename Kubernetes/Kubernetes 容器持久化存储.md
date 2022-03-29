---
title: Kubernetes 容器持久化存储
date: 2022-2-23 16:22:00
tags: [Kubernetes,学习笔记,存储]
category: Kubernetes
---



# 什么是持久化 Volume ？

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



## Local Persistent Volume

适合 Local Persistent Volume 的应用：**高优先级的系统应用，需要在多个不同节点上存储数据，并且对 I/O 较为敏感。比如分布式数据存储等，分布式文件等，需要在本地磁盘上进行大量数据缓存的分布式应用。**

Local Persistent Volume 的应用必须具备数据备份和恢复能力。

> 不能将宿主机的一个目录当作 PV 使用，因为宿主机所在的磁盘可能随时都被应用写满，严重时整个宿主机会直接宕机。 而且，不同的把本地目录直接缺乏 I/O 隔离机制。

所以 Local Persistent Volume 对应的存储介质，一定是一块额外挂载在宿主机的磁盘或者块设备。

可以将 一个 PV 看做一块存储介质盘。

在开始使用 Local Persistent Volume 之前，需要有一个前提条件，需要在集群里配置好磁盘或者块设备。在私有环境下：

1.宿主机挂载并格式化一个可用的本地磁盘，用来充当额外挂载的盘。

2.宿主机上挂载几个 RAM Disk (内存盘）来模拟本地磁盘。



# Container Storage Interface （CSI） 插件开发

CSI 插件体系的设计思想，就是把 Provision 阶段及 Kubernetes 里面一部分存储管理功能，从主干代码中剥离出来，做成了几个单独的组件。这些组件会通过 Watch Api 监听 Kubernetes 里与存储相关的事件变化，比如 PVC 的创建，来执行具体的存储管理动作。



## External Components 组件



### Driver Registrar

负责将插件注册到 kubelet 里面，类似于将可执行文件放在插件目录下。但是在具体实现上，Driver Registrar 需要请求 CSI 插件的 Identity 服务来获取插件信息。



### External Provisioner

负责 Provision 阶段。具体实现中， External Provisioner 监听 APIServer 里的 PVC对象。当一个 PVC 被创建时，它就会调用 CSI Controller 的 CreateVolume 方法，创建对应的 PV。

如果是公有云磁盘，则需要调用公有云API 来创建这个 PV。

因为 CSI 插件是独立的，所以它会自己定义一个单独的 Volume 类型。



### Externa Attacher

负责的“Attach 阶段”。它监听 APIServer 里 VolumeAttachment 对象的变化，因为 **VolumeAttachment 对象**是 Kubernetes 确认一个 Volume 可以进入 “**Attach 阶段**” 的重要标志。如果Kubernetes 确定可以进入 Attach 阶段，那么 **External Attacher** 就会调用 CSI Controller 服务的 ControllerPublish 方法，完成它所对应操作。

但是 Mount 阶段不是 **External Attacher 管辖范围。**



# CSI 插件服务

在 Kuvernetes 中，所有的资源定义（Deployment, Service ..）其实是一个 Object ，可以将 Kubernetes 看作一个 Database ，无论是 Operator 还是 CSI 其核心本质都是不停的 Watch 特定的 Object，一但 kubectl 或者其他客户端 “动了” 这个 Object，我们的对应实现程序就 Watch 到变更然后作出相应的响应；**对于 CSI 编写者来说，这些 Watch 动作已经不必自己实现 [Custom Controller](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#custom-controllers)，官方为我们提供了 [CSI Sidecar Containers](https://kubernetes-csi.github.io/docs/sidecar-containers.html)；**并且在新版本中这些 Sidecar Containers 实现极其完善，比如自动的多节点 HA(Etcd 选举)等。

**所以到迄今为止，所谓的 CSI 插件开发事实上并非面向 Kubernetes API 开发，而是面向 Sidecar Containers 的 gRPC 开发，Sidecar Containers 一般会和我们自己开发的 CSI 驱动程序在同一个 Pod 中启动，然后 Sidecar Containers Watch API 中 CSI 相关 Object 的变动，接着通过本地 unix 套接字调用我们编写的 CSI 驱动。**

 CSI 插件会包含三个服务，对应 **External Components 组件**中三个部分。

## CSI Identity

负责对外暴露这个插件本身的信息：

```go
service Identity {
  // return the version and name of the plugin
  rpc GetPluginInfo(GetPluginInfoRequest)
    returns (GetPluginInfoResponse) {}
  // reports whether the plugin has the ability of serving the Controller interface
  rpc GetPluginCapabilities(GetPluginCapabilitiesRequest)
    returns (GetPluginCapabilitiesResponse) {}
  // called by the CO just to check whether the plugin is running or not
  rpc Probe (ProbeRequest)
    returns (ProbeResponse) {}
}
```



## CSI Controller

定义对 CSI Volume (对应 Kubernetes 里的 PV) 的管理接口。比如：创建和删除 CSI Volume、对 CSI Volume 进行 Attach/Dettach（在 CSI 里，这个操作被叫作 Publish/Unpublish），以及对 CSI Volume 进行 Snapshot 等，它们的接口定义如下所示：

```
service Controller {
  // provisions a volume
  rpc CreateVolume (CreateVolumeRequest)
    returns (CreateVolumeResponse) {}
    
  // deletes a previously provisioned volume
  rpc DeleteVolume (DeleteVolumeRequest)
    returns (DeleteVolumeResponse) {}
    
  // make a volume available on some required node
  rpc ControllerPublishVolume (ControllerPublishVolumeRequest)
    returns (ControllerPublishVolumeResponse) {}
    
  // make a volume un-available on some required node
  rpc ControllerUnpublishVolume (ControllerUnpublishVolumeRequest)
    returns (ControllerUnpublishVolumeResponse) {}
    
  ...
  
  // make a snapshot
  rpc CreateSnapshot (CreateSnapshotRequest)
    returns (CreateSnapshotResponse) {}
    
  // Delete a given snapshot
  rpc DeleteSnapshot (DeleteSnapshotRequest)
    returns (DeleteSnapshotResponse) {}
    
  ...
}
```



## CSI Node

CSI Volume 需要在宿主机上执行的操作，都定义在了 CSI Node 服务里面：

```go
service Node {
  // temporarily mount the volume to a staging path
  rpc NodeStageVolume (NodeStageVolumeRequest)
    returns (NodeStageVolumeResponse) {}
    
  // unmount the volume from staging path
  rpc NodeUnstageVolume (NodeUnstageVolumeRequest)
    returns (NodeUnstageVolumeResponse) {}
    
  // mount the volume from staging to target path
  rpc NodePublishVolume (NodePublishVolumeRequest)
    returns (NodePublishVolumeResponse) {}
    
  // unmount the volume from staging path
  rpc NodeUnpublishVolume (NodeUnpublishVolumeRequest)
    returns (NodeUnpublishVolumeResponse) {}
    
  // stats for the volume
  rpc NodeGetVolumeStats (NodeGetVolumeStatsRequest)
    returns (NodeGetVolumeStatsResponse) {}
    
  ...
  
  // Similar to NodeGetId
  rpc NodeGetInfo (NodeGetInfoRequest)
    returns (NodeGetInfoResponse) {}
}
```

Mount 阶段在 CSI Node 里面的接口，是NodeStageVolume 和 NodePublishVolume 两个接口共同实现的。



# 实践操作

## 实现 Identity Server







# 学习资料

《深入剖析Kubernetes》

[如何编写 CSI 插件]: 	"https://mritd.com/2020/08/19/how-to-write-a-csi-driver-for-kubernetes/"

