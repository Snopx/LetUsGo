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



//TODO









## 学习资料

《深入剖析Kubernetes》