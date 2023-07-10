# MemoryGo
MemoryGo是基于Go实现的K-V内存数据库

功能：

- 支持string、list、hash等数据结构
- AOF持久化以及AOF重写
- 命令开启事务具有`原子性`和`回滚机制`
- 集群模式：使用一致性哈希算法保证数据的一致性
- 并行引擎, 无需担心您的操作会阻塞整个服务器.

## 运行MemoryGO

```shell
git clone git@github.com:xiaodainiao/MemoryGo.git
go run main.go
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/d9f5f1806826481c8ca52d206589101c.png)

Memory默认监听127.0.0.1:6379,可以使用redis-cli 或者NetAssist连接服务器

![在这里插入图片描述](https://img-blog.csdnimg.cn/16e1e656512d49529e5bb05b7764a949.png)

Memory会从redis.conf读取配置文件，包括端口号，是否开启持久化，以及集群节点的IP

![在这里插入图片描述](https://img-blog.csdnimg.cn/69c34860c3d84c429c4bdd413a778752.png)

## 集群模式

Memory支持集群模式，在redis.conf文件中添加下列配置

```go
self 0.0.0.0:6379  //自身节点地址
peers 127.0.0.1:6380, 127.0.0.1:6381  //集群其他节点地址
```

可以直接将文件build打包成可执行文件，node1.conf和node2.conf，进行配置，在本地启动一个双节点集群

![在这里插入图片描述](https://img-blog.csdnimg.cn/03db57530f7042108568013be7bd8998.png)

集群模式下，只要连接任意一台节点就可以访问集群中的所有数据：

```go
redis-cli -p 6380
```

## 项目目录

```shell
├── aof
│   └── aof.go
├── appendonly.aof
├── cluster
│   ├── client_pool.go
│   ├── cluster_database.go
│   ├── com.go
│   ├── del.go
│   ├── flushdb.go
│   ├── keys.go
│   ├── ping.go
│   ├── rename.go
│   ├── router.go
│   └── select.go
├── config
│   └── config.go
├── database
│   ├── command.go
│   ├── db.go
│   ├── echo_database.go
│   ├── keys.go
│   ├── ping.go
│   ├── standalone_database.go
│   └── string.go
├── datastruct
│   └── dict
│       ├── dict.go
│       └── sync_dick.go
├── go.mod
├── go.sum
├── interface
│   ├── database
│   │   └── database.go
│   ├── resp
│   │   ├── conn.go
│   │   └── reply.go
│   └── tcp
│       └── hadler.go
├── lib
│   ├── consistenthash
│   │   └── consistenthash.go
│   ├── logger
│   │   ├── files.go
│   │   └── logger.go
│   ├── sync
│   │   ├── atomic
│   │   │   └── bool.go
│   │   └── wait
│   │       └── wait.go
│   ├── utils
│   │   └── utils.go
│   └── wildcard
│       └── wildcard.go
├── main.go
├── m.exe
├── README.md
├── redis.conf
├── resp
│   ├── client
│   │   └── client.go
│   ├── connection
│   │   └── conn.go
│   ├── handler
│   │   └── handler.go
│   ├── parser
│   │   └── parser.go
│   └── reply
│       ├── consts.go
│       ├── error.go
│       └── reply.go
└── tcp
    ├── echo.go
    └── server.go

```
- 根目录: main 函数，执行入口
- config: 配置文件解析
- interface: 一些模块间的接口定义
- lib: 各种工具，比如logger、同步和通配符

- tcp: tcp 服务器实现
- resp: redis 协议解析器
- datastruct: redis 的各类数据结构实现
  - dict: hash 表
- database: 存储引擎核心
  - server.go: redis 服务实例, 支持多数据库, 持久化, 主从复制等能力
  - database.go: 单个 database 的数据结构和功能
  - router.go: 将命令路由给响应的处理函数
  - keys.go: del、ttl、expire 等通用命令实现
  - string.go: get、set 等字符串命令实现
  - list.go: lpush、lindex 等列表命令实现
  - hash.go: hget、hset 等哈希表命令实现
  - transaction.go: 单机事务实现
- cluster: 集群
  - cluster.go: 集群入口
  - com.go: 节点间通信
  - del.go: delete 命令原子性实现
  - keys.go: key 相关命令集群中实现
  - mset.go: mset 命令原子性实现
  - multi.go: 集群内事务实现
  - rename.go: rename 命令集群实现
  - tcc.go: tcc 分布式事务底层实现
- aof: AOF 持久化实现
