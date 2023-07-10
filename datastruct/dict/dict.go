package dict

/*
  1. 该datastruct就是redis数据结构，首先实现字典，因为redis的核心就是Map
*/

//遍历所有的key value，这样传进一个方法，把方法施加到所有的key value
type Consumer func(key string, val interface{}) bool //如果返回为true的就继续遍历，施加到下一个ket和value上

type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	Put(key string, val interface{}) (result int) //当往redis中Put时，需要返回存了几个字符串
	//类似setex指令，如果redis中没有该key就set进去
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	//返回删掉几个元素
	Remove(key string) (result int)
	//遍历整个字典，传入一个方法
	ForEach(consumer Consumer)
	Keys() []string                        //列出所有的键
	RandomKeys(limit int) []string         //随机返回一些键
	RandomDistinctKeys(limit int) []string //随机返回不重复的键
	Clear()
}
