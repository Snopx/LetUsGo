---
title: Kubernetes Job与CronJob
date: 2022-1-20 21:36:00
tags: [Kubernetes,学习笔记]
category: Kubernetes
---



## Batch Job （计算业务）

计算业务会在计算完成后就直接退出了，如果使用 Deployment 来管理这种业务会被 Deployment Controller 不断重启。



**定义 Job API 对象：**

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: resouer/ubuntu-bc 
        command: ["sh", "-c", "echo 'scale=10000; 4*a(1)' | bc -l "]
      restartPolicy: Never
  backoffLimit: 4
$ kubectl create -f job.yaml
```

注意：离线计算的 Pod 永远都不应该被重启，否则它们会再重新计算一遍。

所以 Kubernetes限制了 restartPolicy 在 Job 对象里面的值：只允许设置为: Never 和 OnFailure。



**如果离线作业失败了要怎么办？**

可以定义 restartPolicy = Never，那么离线作业失败后 Job Controller 就会不断地尝试创建一个新 Pod。

这个尝试不能无限制地进行下去，用来限制重试次数的字段名为：**backoffLimit 可以设置该字段值来限制重试次数（默认为6次）**

spec.activeDeadlineSeconds 字段可以设置最长运行时间，一旦运行超过100 s，这个Job 的所有

```yaml
spec:
 backoffLimit: 5
 activeDeadlineSeconds: 100
```



## Job Controller 对并行作业的控制方法

在 Job 对象中，负责并行控制的参数有两个：

1.spec.parallelism ，它定义的是一个 Job 在任意时间最多可以启动多少个 Pod 同时运行

2.spec.completions，它定义的是 Job 至少要完成的 Pod 数目，即 Job 的最小完成数。

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  parallelism: 2       //最大并行数为 2
  completions: 4       //最小的完成数是 4
  template:
    spec:
      containers:
      - name: pi
        image: resouer/ubuntu-bc
        command: ["sh", "-c", "echo 'scale=5000; 4*a(1)' | bc -l "]
      restartPolicy: Never
  backoffLimit: 4
```

Job Controller 控制的对象，直接就是 Pod，在控制循环中进行 Reconcile 操作，是根据实际在 Running 状态 Pod 的数目，已经成功退出的 Pod 的数目，以及 parallelism，completions 参数的值共同计算出在这个周期里，应该创建或者删除的 Pod 数目，然后调用 Kubernetes API 来执行这个操作。

Job Controller 实际上控制了，作业执行的并行度，以及总共需要完成的任务数这两个重要参数。而在实际使用时，需要根据作业的特性，来决定并行度和任务书的合理取值。



## CronJob

```yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
```

**CronJob 是一个 Job 对象的控制器（Controller），它创建和删除 Job 的依据，是 schedule 字段定义的，一个标准的 Unix Cron 格式的表达式。**

由于定时任务的特殊性，很可能某个 Job 还没执行完，另一个新 Job 就产生了。可以通过 spec.concurrencyPolicy 字段来定义具体的处理策略。

Allow 默认，Job 可以同时存在。

Forbid，不会创建新的 Pod ，改创建周期会被跳过。

Replace，这意味着新产生的 Job 会替换旧的，没有执行完的 Job。

**如果某一次 Job 创建失败，这次创建会被标记为 Miss 。在指定的时间内，Miss 次数到达100 时，CronJob 会停止再创建这个 Job。**

可以通过 spec.startingDeadlineSeconds 字段设置指定时间，例如设置为 300 秒，或更多。



## 学习资料

《深入剖析Kubernetes》