package reply

type PongReply struct {
}

var pongbytes = []byte("+PONG\r\n") //客户端发送ping,服务端发送pong

//实现reply接口的方法
func (r PongReply) ToBytes() []byte {
	return pongbytes
}

/*外面的方法想要一个PongReply时，不需要new一个而是直接调用Make
优点：就是自己在自己包里修改实现，或者将PongReply结构体改为私有的
方便外面调用*/

/*
MakePongReply() 函数的作用是创建一个新的 PongReply 对象，并返回其指针。
它的存在是为了提供一种简便的方式来创建 PongReply 对象，
而不是直接使用 new(PongReply) 或 &PongReply{} 这样的方式。
*/
func MakePongReply() *PongReply {
	return &PongReply{} //相当于初始化了
}

//======================== OkReply is +OK======================
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes marshal redis.Reply
func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply)

// MakeOkReply returns a ok reply
func MakeOkReply() *OkReply {
	return theOkReply
}

//=======================空字符串回复=====================
var nullBulkBytes = []byte("$-1\r\n")

// NullBulkReply is empty string
type NullBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply creates a new NullBulkReply
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

//======================空数组回复=========================
var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply is a empty list
type EmptyMultiBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

//======================= NoReply respond nothing, for commands like subscribe
type NoReply struct{}

var noBytes = []byte("")

// ToBytes marshal redis.Reply
func (r *NoReply) ToBytes() []byte {
	return noBytes
}
