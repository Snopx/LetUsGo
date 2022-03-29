---
title: Kubernetes 调度控制原理
date: 2022-3-24 13:54:00
tags: [Kubernetes,学习笔记,调度]
category: Kubernetes
---

# Kubernetes 调度控制原理

API Server 作为集群的核心，负责集群各功能模块之间的通信。集群内的功能模块通过 API Server 将信息存入 etcd ,其他模块通过 API Server 读取这些信息，从而实现模块之间的信息交互。

为了缓解集群各模块对 API Server 的访问压力，各功能模块都采用缓存机制来缓存数据。各功能模块定时从 API Server 获取指定资源对象信息，然后将这些信息保存到本地缓存，功能模块在某些情况下不直接访问 API Server ，而是通过访问缓存数据来间接访问 API Server。



## 调度控制原理

Controller Manager 作为集群内部的管理控制中心，负责集群内的 Node , Pod 副本，服务端点，命名空间，服务账号，资源定额等的管理并执行自动化修复流程，确保集群处于预期的工作状态。

在 Kubernetes 集群中，每个 Controller 就是一个操作系统，它通过 API Server 监控系统的共享状态，并尝试将系统从”现有状态“修正到”期望状态“。



## Replication Controller

Replication Controller （副本控制器） 的核心作用是确保任何时候集群中一个 RC （资源对象） 所关联的 Pod 都保持一定数量的 Pod 副本处于正常运行状态。

**不要越过 RC 直接创建 Pod 。**

Service 可能被不同 RC 管理的多个 Pod 副本组成，在 Service 的整个生命周期里，由于需要发布不同版本的 Pod，因此希望不断有旧的 RC 被销毁，新的 RC 被创建。Service 自身及其他的客户端应该不需要关注 RC。

Pod 实例都是通过 RC 里定义的 Pod 模板（Templet）创建的，改模板包含 Pod 的标签属性，同时 RC 里包含一个标签选择器（Label Selector）, Selector 的值表明了该 RC 所关联的 Pod。RC 会保证每个由它创建的 Pod 都包含与它的标签选择器相匹配的 label 。

**通过 RC 创建的 Pod 副本在初始阶段状态是一致的。**



### Pod 模板

模板就像一个模具，模具制作出来的东西一旦离开了模具，它们之间就再也没有关系了。同样，Pod 被创建以后，模具再发生变化，也不会影响到已经创建的 Pod。

Pod 可以通过修改它的标签脱离 RC 的管控。可以将 Pod 从集群中迁移，数据修复等调试操作。删除一个 RC 不会影响它所创建的 Pod。



### Replication Controller 的作用

确保当前集群中只有 N 个 Pod 实例，N 是 RC 中定义的 Pod 副本数量。

通过调整 RC 的 spec.replicas 属性值来调整 Pod 的副本数量。

副本控制器常用模式：

1.重新调度（Rescheduling）,副本控制器能确保指定数量的副本存在集群中，即使发生节点故障或者被终止等状况。

2.弹性伸缩（Scaling）, 手动或者通过自动扩容代理修改副本控制器的属性一般使用：

```
kubectl scale --replicas=3 replicationcontrllers rcName
```

3.滚动更新（Rolling Updates），副本控制器被设计成通过逐个替换 Pod 的方式来辅助服务的滚动更新。推荐的方式是创建一个新的只有一个副本的 RC ，若新的 RC 副本数量加1，则旧的 RC 副本数量减1，直到旧的 RC 副本为 0，然后删除旧的 RC 。

```
kubectl rolling-update frontend-v1 -f frontend-v2.json
```



## Node Controller 

Node Controller 负责发现，管理和监控集群中的各个 Node 节点。Kubelet  启动时通过 API Server 注册节点信息，并定时向 API Server 发送节点信息。API Server 接受后，会将这些信息写入 etcd 。

1.Controller Manager 在启动时如果设置了 --cluster-cidr 参数，那么为每个没有设置 Spec.PodCIDR  的 Node 节点生成一个 CIDR 地址，并且用 CIDR 地址设置节点的 Spec.PodCIDR 属性，以防止不同节点的 CIDR 地址发生冲突。

2.逐个读取节点信息，用 Node Controller 所在节点的系统时间作为探测时间，用上次节点信息中的节点状态变化时间作为该节点的状态变化时间。

如果判断一段时间没有收到节点的状态信息，将改变节点状态，并通过 API Server 保存节点状态。

3.逐个读取节点信息，如果节点状态为 ”非就绪“，则将节点加入待删除队列，否则将节点从该队列中删除。如果节点状态为”非就绪“状态，且系统指定了 Cloud Provider ，则 Node Controller 调用 Cloud Provider 查看节点，若发现节点故障，则删除 etcd 中的节点信息，并删除和该节点相关的 Pod 等资源的信息。



## ResourceQuota Controller (资源配额管理)

资源配额管理确保了指定的对象在任何时候都不会超量占用系统资源，避免了某些业务进程设计或者实现的缺陷导致整个系统运行紊乱甚至意外宕机，对整个集群的平稳运行和稳定性有非常重要的作用。

Kubernetes 支持三个层次的资源配额管理。

1.容器级别，可以对 CPU 和 Memory 进行限制。

2.Pod 级别，对 Pod 中所有容器的可用资源进行限制。

3.Namespace 级别，为 Namespace  级别的资源限制，包括 Pod 数量，Replication Controller 数量， Service 数量，ResourceQuota 数量，Secret 数量，可持有的 PV 数量。



### ResourceQuota Controller 

负责实现 Kubernetes 的资源配额管理。可通过 API Server 为 Namespace 维护 ResourceQuota  对象，API Server 将该对象保存到 etcd 中。所有的资源对象的实时状态都保存在 etcd 中，方便 ResourceQuota Controller 在计算资源使用总量。

ResourceQuota Controller 以 Namespace 作为分组统计单元，通过 API Server 定时读取 etcd 中每个 Namespace 里定义的 ResourceQuota 信息，计算 Pod ，Service  等资源对象的总数，以及所有 Container 实例使用的资源量。然后会将统计结果写入 etcd z中。

所以当用户创建或修改资源时，API Server 会调用 ResourceQuota  插件，然后读取配额统计的结果，由此来判断某个资源的配额是否允许操作正常执行，如果已使用完，则拒绝该请求。



## Namespace Controller

Namespace Controller 能控制 Namespace 的创建和删除，并且它会时刻观察 Namespace 设置的删除期限，同时如果该 Namespace 的 spec.inalizers 域值是空的，那么 Namespace Controller 会通过 API Server 删除该 Namespace 资源。

Namespace Controller 删除某一个 Namespace 时，同时会删除该 Namespace 下的 ServiceAccount ， RC , Pod ，Secret ，PersistentVolume , ListRange , ResouceQuota 和 Event 等资源对象。



## ServiceAccount Controller 与 Token Controller

ServiceAccount Controller 与 Token Controller 是与安全相关的两个控制器。

ServiceAccount Controller 监听 Service Account 的删除事件和 Namespace 的创建，修改事件。如果在 Namespace 中没有 deault Service Account 会自动创建一个。



Token Controller 对象监听  Service Account 的创建，修改和删除事件。

1.如果监听的事件是创建和修改 Service Account 事件，则读取 Service Account 信息；

2.如果 Service Account 没有 Service Account Secret (用于访问 API Server 的 Secret)，则创建一个 **JWT Token，并会建立相关连接。**

3.如果监听的事件是是删除 Service Account，则删除该 Service Account 相关的 Secret。

Token Controller 同时监听 Secret 的创建，删除，修改事件。它用于控制 Secret 对象。



## Service Controller 与 Endpoint Controller

Kuvernetes 中的 Service 是一种资源对象，和 Pod 相似，Kubernetes 指派一个集群 IP 给该 Service。

Endpoint 资源包含一个 地址和端口的集合。这些 IP 地址和端口号即通过标签选择器过滤出来的 Pod 的访问端点。

如果是一个不带标签选择器的 Service ，系统不会自动创建 Endpoint ,需要手动创建它，用于指向实际的后端访问地址。如创建一个数据库 Service。



**如何通过虚拟 IP 访问到后端 Pod ?**

Kubernetes 集群中每个节点上都运行一个 “kube-proxy” 的进程，该进程监控 Master 节点添加和删除 “Service” 和 “Endpoint” 的行为。

Kube-proxy 为每个 Service 在本地主机上随机开一个端口，任何访问这个端口的连接都被代理到一个后端 Pod 上。Kube-proxy 根据 **Round Robin** 算法及 Service 的 Session 粘连（SessionAinity）决定选择哪个后端。

Kube-proxy 在本机的 Iptables 中安装规则，这些规则会让流量重定向到**随机端口，再**通过该端口流量再被 kube-proxy 转到相应的后端 Pod 上。

服务 Endpoint  模型会创建后端 Pod 和 IP 和端口列表（包含在 Endpoints 对象中），Kube-proxy  就是从这个 Endpoint 列表中选择服务后端的。**集群内的节点通过虚拟 IP 和端口能发访问 Service 后端的 Pod。**

> 默认情况下，Kubernetes 会为 Service 指定一个集群 IP。 如果需要给某个 Service 指定集群 IP ,可以在 Service 的 spec.clusterIP 域中设置所需要的 IP 地址。 为 Service 指定的 IP 地址必须在集群的 CIDR 范围内，如果 IP 地址违法，那么 API Server 会返回 422 状态码

Kubernetes 支持两种主要模式来找到 Service

**1.容器的 Service 环境变量**

创建一个 Pod 时， Kubelet 在该 Pod 中所有容器中为当前所有 Service 添加一系列环境变量。

通过环境变量会带来一个不好的结果：任何被某个 Pod 所访问的 Service，必须先于该 Pod 被创建。否则和这个后创建的 Service 相关的环境变量，将不会被加入该 Pod 的容器中。

**2.DNS 寻找服务**

DNS 服务器通过 Kubernetes API 监控与 Service 相关的动作。当添加 Service 时， DNS 服务器为每个 Service 创建一系列 DNS 记录。如果其他 Namespace 访问这个 Service 则通过 “DNS Service Name.Namespace” 来查找这个 Service。DNS 返回的是集群 IP （虚拟 IP，Cluster IP）

Kubernetes 也支持 DNS SRV 被命名端口记录。

Service 也能提供一个供集群外部用户访问的 IP 地址，甚至是公网 IP 地址，通过这个 IP 来访问集群的 Service。

Kubernetes 提供两种方式来满足以上需求。

**ClusterIP 默认值，仅使用集群内部虚拟 IP （集群IP，Cluster IP）**

**NodePort 使用虚拟 IP（集群IP，Cluster IP），同时通过在每个节点上暴露相同的端口来暴露 Service。**

**LoadBalancer 使用虚拟 IP（集群IP，Cluster IP）和 NodePort ，同时请求云服务商作为转向 Service 的负载均衡器。**



# Kubernetes Scheduler

Kubernetes Scheduler 的作用是将待调度的 Pod 按照特定的调度算法和调度策略绑定到集群中某个 Node 上，并将绑定信息写入 etcd  中。

Kubernetes Scheduler 默认调度流程分为两步：

1.预选调度过程，即遍历所有目标 Node ,筛选出符合要求的候选节点。（Kubernetes 内置多种预选策略供用户选择）。

2.确定最优的节点，在第一步的基础上，采用优选策略计算出每个候选节点的积分，积分最高的就选择它。

Scheduler 预选策略包含：NoDiskConflict , PodFitsResources , PodSlelectorMatches , PodFitsHost , CheckNodeLabelPresence ,

CheckServiceAffinity 和 PodFitsPorts 策略等。每个节点只有通过这五个默认预选策略后才能初步被选中。



## Kubelet 运行机制分析

Kubernetes 集群中，在每个 Node 节点上都会启动一个 Kubelet 服务进程。该进程用于处理 Master 节点下发到本节点的任务，管理 Pod 及 Pod 中的容器。每个 Kubelet 进程会在 API Server 上注册节点自身信息，定期向 Master 节点汇报节点资源使用情况，并通过 cAdvise 监控容器和节点资源。

### 节点管理

参数设置：

--register-node :是否向 API Server 注册自己。

--api-servers : 告诉 API Server 的位置。

--kubeconfig : 告诉 kubelet 哪里可以找到用于访问 API Server 的证书

--cloud-provider: 告诉 kubelet 如何从云服务商读取到和自己相关的元数据。

--node-status-update-frequency 设置 kubelet 每隔多少时间向 API Server 报告节点状态，默认为 10 秒。



### Pod 管理

Kubelet 监听 etcd ，所有针对 Pod 的操作都会被监听到。如果有创建/修改/删除等操作则直接执行。

Kubelet 创建和修改 Pod 过程如下：

1.为该 Pod 创建一个数据目录

2.从 API Server 读取该 Pod 清单

3.为该 Pod 挂载外部卷（Extenal Volume）

4.下载 Pod 用到的 Secret

5.检查已经运行在节点中的 Pod ，如果 Pod 没有容器或者 Pause 容器没有启动，则先停止 Pod 中所有容器的进程。如果在 Pod 中有需要删除的容器，则删除这些容器。

6.pause 镜像为每个 Pod 创建一个容器。Pause 容器用于接管 Pod 中所有其他容器的网络。每创建一个新的 Pod ，Kubelet 都回创建一个 Pause 容器，然后再创建其他容器。

7.为 Pod 中每个容器计算一个 hash 值，然后用容器的名字去 Docker 查询对应容器的 hash 值。

1.如果两者的值不同，则停止 Docker 中容器的进程，并停止与之关联的 Pause 容器的进程

2.如果两者相同，不做任何处理

3.如果容器被终止了，并没有任何的重启策略，则不做任何处理

调用 Docker Client 下载容器镜像，调用 Docker Client 运行容器。



### 容器健康检查

**LivenessProbe 探针**，用于判断容器是否健康：

1.如果容器不健康，Kubelet 将删除该容器，并根据容器的重启策略做出相应的操作。

2.如果不包含 LivenessProbe 探针，则返回值永远是 "Success"

LivenessProbe 包含三种实现方式来检测：

1. ExecAction 执行命令，看命令退出状态是否为 0
2. TCPSocketAction 通过IP 地址和端口号执行 TCP 检查，如果端口能访问，则证明容器健康。
3. HTTPGetAction   通过IP 地址和端口号调用 Http Get 请求，通过判断响应状态码来判断容器是否健康

**ReadinessProbe 探针**，用于判断容器是否启动完成，且准备接受请求。如果检测到失败，则 Pod 的状态将被修改。 Endpoint Controller 将从 Service 的 Endpoint 中删除包含容器所在 Pod 的 IP 地址的 Endpoint 条目。





# 学习资料

《Kubernetes 权威指南》
