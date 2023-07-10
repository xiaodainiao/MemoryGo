package cluster

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

//rename没有转发，目前没实现rename在不同节点的功能，可能一改就取到其他的节点（可以将k1删掉，然后再插入key2）
/*
1. 首先判断修改完后还会在同一个节点吗如果是的话就rename，否则就不改
2. rename k1 k2
*/
// Rename renames a key, the origin and the destination must within the same node
func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	//获取改过还源地址的hash看看是否相等，如果相等则是在同一个节点上
	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one slot in cluster mode")
	}
	return cluster.relay(srcPeer, c, args)
}
