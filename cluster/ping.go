package cluster

import "xiaodainiao/interface/resp"

//func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {}
//本地执行
func ping(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdAndArgs)
}
