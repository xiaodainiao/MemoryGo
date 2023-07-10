package cluster

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

func flushdb(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs) //获取所有节点返回结果，遍历返回结果只有当全部ok时才是ok否则就return错误
	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrorReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOkReply()
	}
	return reply.MakeErrReply("error:" + errReply.Error())

}
