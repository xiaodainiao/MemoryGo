package resp

type Connection interface {
	Write([]byte) error //给客户端回复消息
	GetDBIndex() int    //因为Redis有16个DB，每一个DB的key和value都是隔离开的，运行时首先选择第几号DB
	SelectDB(int)       //DB之间进行切换
}
