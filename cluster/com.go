package cluster

import (
	"context"
	"errors"
	"strconv"
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/utils"
	"xiaodainiao/resp/client"
	"xiaodainiao/resp/reply"
)

/*
负责节点与节点之间通信
从连接池取出一个连接与其他节点通信，通信完成后放回连接
*/

//从连接池拿到一个连接，用来转发指令，入参就是peer的ip
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	//从ObjectPool中拿到属于该peer的连接，然后再从连接池中取出一个连接

	factory, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection factory not found")
	}
	raw, err := factory.BorrowObject(context.Background()) //从连接池获取一个连接
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(*client.Client) //BorrowObject返回一个空接口，使用类型断言转成我们需要的Client
	if !ok {
		return nil, errors.New("connection factory make wrong type")
	}
	return conn, nil
}

//返回连接到连接池，否则会导致连接耗用
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}

//请求转发
//resp.Connection记录用户底层的connection使用的哪个DB,args指令
// relay relays command to peer
// select db by c.GetDBIndex()
// cannot call Prepare, Commit, execRollback of self node
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self { //如果是自己的节点就直接执行get/set不需要转发
		// to self db
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	//避免连接耗尽注册defer归还连接
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	//调用客户端的Send
	//client发送指令它发送的几号DB(1-16)是记录在本地的，而兄弟节点不知道是几号DB，所以兄弟节点永远是0号DB
	//因此先给兄弟节点发送一个Select指令
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	return peerClient.Send(args)
}

// broadcast broadcasts command to all node in cluster //广播是一组reply
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	result := make(map[string]resp.Reply) //广播是一组reply
	for _, node := range cluster.nodes {
		reply := cluster.relay(node, c, args)
		result[node] = reply
	}
	return result
}
