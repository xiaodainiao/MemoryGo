package database

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

// Ping the server，
func Ping(db *DB, args [][]byte) resp.Reply {
	//if len(args) == 0 {
	//	return &reply.PongReply{}
	//} else if len(args) == 1 {
	//	return reply.MakeStatusReply(string(args[0]))
	//} else {
	//	return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	//}
	return reply.MakePongReply()
}

//把ping注册到cmTable上，ping是一个指令
func init() { //这个init保证这个包在调用时就调用init方法
	RegisterCommand("ping", Ping, -1)
}
