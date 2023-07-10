package cluster

import (
	"context"
	"fmt"
	"github.com/jolestar/go-commons-pool/v2"
	"runtime/debug"
	"strings"
	"xiaodainiao/config"
	"xiaodainiao/database"
	databaseface "xiaodainiao/interface/database"
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/consistenthash"
	"xiaodainiao/lib/logger"
	"xiaodainiao/resp/reply"
)

//集群层/转发层
// ClusterDatabase represents a node of godis cluster
// it holds part of data and coordinates other nodes to finish transactions
type ClusterDatabase struct {
	self string //自己的名称，自己的地址

	nodes          []string //存放整个集群的节点
	peerPicker     *consistenthash.NodeMap
	peerConnection map[string]*pool.ObjectPool //连接池string为a或者b节点的地址//保存多个连接池例如a要保存b,c的连接池使用map保存
	//连接池可以自动建立连接，你要告诉我怎末建立做一个工厂
	db databaseface.Database //cluster_database下层就是standalone_database（记录一下下层）
}

// MakeClusterDatabase creates and starts a node of cluster
//使用连接池
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self: config.Properties.Self,

		db:             database.NewStandaloneDatabase(), //集群调用单机版的database
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1) //nodes记录自己和双方的所有节点
	//将Peers和self都放进该数组
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.peerPicker.AddNode(nodes...) //将所有的节点进行一致性哈希
	ctx := context.Background()
	for _, peer := range config.Properties.Peers { //对每一个peers的ip启动一个连接工厂
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster
}

// CmdFunc represents the handler of a redis command
//定义一个方法的声明，所有节点过来都要查询指令具体执行方法在哪里（执行方法里面就是具体的转发，群发还是本地）
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply

// Close stops current node of cluster
func (cluster *ClusterDatabase) Close() {
	cluster.db.Close() //关闭单击版db
}

var router = makeRouter() //调用路由表

// Exec executes command on cluster
//集群层执行相当于替换了单机版的执行，也就是指令解析完后首先调用它
func (cluster *ClusterDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(cmdLine[0])) //先识别是什么指令
	cmdFunc, ok := router[cmdName]                 //找到对应的CmdFunc
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, cmdLine)
	return
}

// AfterClientClose does some clean after client close connection
func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
