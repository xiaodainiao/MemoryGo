package resp

/*各种回复*/
type Reply interface {
	ToBytes() []byte //因为TCP是字节流的，所以需要把发送的数据全部转换为字节
}
