package cluster

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

//del k1 k2 k3 k4广播删除
/*
可能遇到该key没在同一个节点，而且del删除返回的是删除了几个
因此需要把所有的回复汇总，把最终的结果分析
*/
// Del atomically removes given writeKeys from cluster, writeKeys can be distributed on any node
// if the given writeKeys are distributed on different node, Del will use try-commit-catch to remove them
func Del(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply reply.ErrorReply
	var deleted int64 = 0 //回复删除了几个key
	for _, v := range replies {
		if reply.IsErrorReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
		intReply, ok := v.(*reply.IntReply) //类型转换为Int
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deleted += intReply.Code
	}

	if errReply == nil {
		return reply.MakeIntReply(deleted)
	}
	return reply.MakeErrReply("error occurs: " + errReply.Error())
}
