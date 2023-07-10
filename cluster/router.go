package cluster

import "xiaodainiao/interface/resp"

//做一个map存放存放指令和该指令的执行方法的各种实现之间的关系

//格式:type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply
// relay command to responsible peer, and return its reply to client
//Get key //Set k1 v1
/*
1. 首先找到需要转发给那个节点（用一致性哈希找到该key）
*/
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key) //返回节点ip
	return cluster.relay(peer, c, args)      //转发
}

//key:指令,value:CmdFunc 路由表
func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["ping"] = ping

	routerMap["del"] = Del

	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename
	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc
	routerMap["select"] = execSelect
	routerMap["flushdb"] = FlushDB

	return routerMap
}
