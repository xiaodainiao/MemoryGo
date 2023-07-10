package database

import "xiaodainiao/interface/resp"

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

// Database is the interface for redis style storage engine
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection) //关闭后的善后工作，可能客户端关闭后，需要删除一些数据
	Close()
}

//指代redis数据结构，它返回一个interface可以指代所有类型，set/hash/zset/list/string等
// DataEntity stores data bound to a key, including a string, list, hash, set and so on
type DataEntity struct {
	Data interface{}
}
