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

![image-20230710160059633](C:\Users\Administrator\AppData\Roaming\Typora\typora-user-images\image-20230710160059633.png)

Memory默认监听127.0.0.1:6379,可以使用redis-cli 或者NetAssist连接服务器

![image-20230710160534011](C:\Users\Administrator\AppData\Roaming\Typora\typora-user-images\image-20230710160534011.png)

Memory会从redis.conf读取配置文件，包括端口号，是否开启持久化，以及集群节点的IP

![image-20230710160815944](C:\Users\Administrator\AppData\Roaming\Typora\typora-user-images\image-20230710160815944.png)

## 集群模式

Memory支持集群模式，在redis.conf文件中添加下列配置

```go
self 0.0.0.0:6379  //自身节点地址
peers 127.0.0.1:6380, 127.0.0.1:6381  //集群其他节点地址
```

可以直接将文件build打包成可执行文件，node1.conf和node2.conf，进行配置，在本地启动一个双节点集群

<img src="C:\Users\Administrator\AppData\Roaming\Typora\typora-user-images\image-20230710161057147.png" alt="image-20230710161057147"  />

![image-20230710161116228](C:\Users\Administrator\AppData\Roaming\Typora\typora-user-images\image-20230710161116228.png)

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

