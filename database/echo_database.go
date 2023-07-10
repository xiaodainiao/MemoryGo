package database

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

//==================测试什么也没做========================
type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

//调用exec说明用户发送的TCP报文已经被解析成为指令了，解析完指令在把它包装成TCP报文，发送回去
func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	//TODO implement me
	panic("implement me")
}

func (e EchoDatabase) Close() {
	//TODO implement me
	panic("implement me")
}
