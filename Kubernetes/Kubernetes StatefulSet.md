---
title: Kubernetes StatefulSet
date: 2022-1-17 20:35:00
tags: [Kubernetes,学习笔记]
category: Kubernetes
---



## Kubernetes StatefulSet

为了支持 “有状态应用”，Kubernetes 扩展出一个编排功能：StatefulSet

它把真实世界的应用状态抽象为两种情况：

1.**拓扑状态**。这种情况意味着，应用的多个实例之间不是完全对等的关系。这些应用实例，必须按照某些顺序启动，比如应用的主节点 A 要先于从节点 B 启动。而如果你把 A 和 B 两个 Pod 删除掉，它们再次被创建出来时也必须严格按照这个顺序才行。并且，新创建出来的 Pod，必须和原来 Pod 的网络标识一样，这样原先的访问者才能使用同样的方法，访问到这个新 Pod。

2.**存储状态**。这种情况意味着，应用的多个实例分别绑定了不同的存储数据。对于这些应用实例来说，Pod A 第一次读取到的数据，和隔了十分钟之后再次读取到的数据，应该是同一份，哪怕在此期间 Pod A 被重新创建过。这种情况最典型的例子，就是一个数据库应用的多个存储实例。

**所以，StatefulSet 的核心功能，就是通过某种方式记录这些状态，然后在 Pod 被重新创建时，能够为新 Pod 恢复这些状态。**



## Headless Service

Service 是 Kubernetes 项目中用来将一组 Pod 暴露给外界访问的一种机制。比如可以通过 Service 它能访问到某个具体的 Pod 。

**Service 的 VIP(Virtual IP, 虚拟 IP) 方式**

通过访问 Service 的虚拟 IP ,它会把请求转发到该 Service 所代理的某一个 Pod 上。

**Service 的 DNS 方式**

只要访问 DNS 记录可以访问到 Service 所代理的某一个 Pod。

Headless Service 不需要分配一个 VIP，而是可以直接以DNS 记录的方式解析出被代理 Pod 的 IP 地址。

StatefulSet YAML 文件解读

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"    
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.9.1
        ports:
        - containerPort: 80
          name: web
```

**serviceName 字段**

这个字段是为了告诉 StatefulSet 控制器在执行控制循环的时候请使用 nginx 这个 Headless Service 来保证 Pod 的 “可解析身份”

> StatefulSet 会给它所管理的所有 Pod 的名字进行编号，而且全都是从0开始累加，与 StatefulSet 的每个 Pod 实例一一对应，绝不重复。 即使删除 Pod ，会按照原先编号的顺序，创建出两个新的 Pod，并且它们的网络信息与旧 Pod 相同。

但是要注意，尽管网络记录不回发生改变，但是解析到的 Pod 的 IP 地址并不是固定的。这意味着我们对 “有状态应用” 实例的访问必须使用 DNS 记录或 hostname 的方式，而不应该直接访问这些 Pod 的 IP 地址。



## Persistent Volume Claim

Kubernetes 项目引入了一组叫做 Persistent Volume Claim (PVC) 和 Persistent Volume （PV）的 API 对象，降低了用户声明和使用持久化 Volume 的门槛。

定义一个 PVC ，申明想要的 Volume 属性：

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pv-claim
spec:
  accessModes:
  - ReadWriteOnce     // 描述模式（表示挂载方式是可读写的，并且只能挂在在一个节点上）
  resources:
    requests:
      storage: 1Gi    //描述想要的 Volume 大小
```

在应用的 Pod 中使用这个 PVC：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pv-pod
spec:
  containers:
    - name: pv-container
      image: nginx
      ports:
        - containerPort: 80
          name: "http-server"
      volumeMounts:
        - mountPath: "/usr/share/nginx/html"
          name: pv-storage
  volumes:
    - name: pv-storage
      persistentVolumeClaim:
        claimName: pv-claim
```

可以看到 Pod 的 Volumes 定义中，只需要声明它的类型是 persistentVolumeClaim，然后指定 PVC的名字，不用关心 Volume 本身。

> 符合条件的 Volume  来自运维人员维护的 PV（Persistenrt Volume）对象

当把一个 Pod 删除后，这个 Pod 对应的 PVC 和 PV ，并不会被删除，而这个 Volume 里已经写入的数据也依然会保存在远程存储服务里。



## 小结

SatatefulSet 控制器直接管理的是 Pod 。这是因为 StatefulSet 里的不同 Pod 实例，不再像 ReplicaSet 中完全一样，而是都存在细微区别。

Kubernetes 通过 Headless Service 为有编号的 Pod ，在 DNS 服务器中生成带有同样编号的 DNS 记录。这是为了控制网络访问。

StatefulSet 还为每一个 Pod  分配并创建一个同样编号的 PVC ，这样 Kubernetes 可以通过 Persistent Volume 机制为这个 PVC 绑定上对应的 PV，从而保证了每一个 Pod 都拥有一个独立的 Volume。

StatefulSet 其实就是一种特殊的 Deployment 。它的每个 Pod 都有一个编号，这个编号会体现出 Pod 的名字和 hostname 等标识信息，这不仅代表 Pod 的创建顺序，也是 Pod 的重要网络标识。



## 学习资料

《深入剖析Kubernetes》