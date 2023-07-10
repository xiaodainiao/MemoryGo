package reply

import (
	"bytes"
	"xiaodainiao/interface/resp"

	"strconv"
)

/*一般的回复，除了常量回复和错误回复之外的动态回复
服务端向客户端回复的内容*/

var (
	nullBulkReplyBytes = []byte("$-1")

	// CRLF is the line separator of redis serialization protocol
	CRLF = "\r\n" //结尾，因为RESP协议最后都是\r\n
)

//把前面xiaodainiao转换为$11\r\nxiaodainiao\r\n
type BulkReply struct {
	Arg []byte //把它变成适合Reids协议的回复比如回复"xiaodainiao" 协议上$11\r\nxiaodainiao\r\n
}

// ToBytes marshal redis.Reply
func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	//把int转换为字符串
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

//=============================字符串数组回复=====================================
type MultiBulkReply struct {
	Args [][]byte
}

// ToBytes marshal redis.Reply
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer //更适合做字符的拼装
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args { //遍历所有字符串
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

// MakeMultiBulkReply creates MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

//==========================回复通用状态======================
// StatusReply stores a simple status string
type StatusReply struct {
	Status string
}

// MakeStatusReply creates StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

// ToBytes marshal redis.Reply
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF) //+OK\r\n或者其他状态
}

//===========================通用数字回复=======================
// IntReply stores an int64 number
type IntReply struct {
	Code int64
}

// MakeIntReply creates int reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// ToBytes marshal redis.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

//=========================一般错误回复（动态的）--自定义状态回复=======================
// ErrorReply is an error and redis.Reply
type ErrorReply interface { //给错误的回复写接口
	Error() string
	ToBytes() []byte
}

// StandardErrReply represents handler error
type StandardErrReply struct {
	Status string
}

// ToBytes marshal redis.Reply
func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

// MakeErrReply creates StandardErrReply
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// 判断是正常回复还是异常回复，就是判断第一个符号是不是-
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
