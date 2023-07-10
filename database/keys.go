package database

import (
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/utils"
	"xiaodainiao/lib/wildcard"
	"xiaodainiao/resp/reply"
)

//实现DEL EXISTS KEYS FLUSHDB TYPE RENAME RENAMENX
// execDel removes a key from db
//DEL K1 K2 K3之前db中的Exec已经把前面的DEL干掉了，所以传进来的只有3个参数
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}

	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

//EXISTS K1 K2 K3...查询哪个存在
// execExists checks if a is existed in db
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// execFlushDB removes all data in current db
//FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	//return &reply.OkReply{}
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOkReply() //都要返回一个实现Reply接口的
}

//TYPE->键的类型就是args[][]中的第一个成员
// execType returns the type of entity, including: string, list, hash, set and zset
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key) //返回Data interface{}
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	//TODO:后面实现SET/ZSET/LIST等
	return &reply.UnknownErrReply{}
}

// execRename a key
//将KEY修改名字RENAME K1 K2, 把K1:v删除改成K2:v
func execRename(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])  //取出K1
	dest := string(args[1]) //取出K2

	entity, ok := db.GetEntity(src)
	if !ok { //不存在返回错误
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("rename", args...))
	return &reply.OkReply{}
}

//RENAMENX K1 K2可能k2也存在，把k1改成k2此时k2相当于删掉了
// execRenameNx a key, only if the new key does not exist
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.Removes(src, dest) // clean src and dest with their ttl
	db.addAof(utils.ToCmdLine2("renamenx", args...))
	db.PutEntity(dest, entity)
	return reply.MakeIntReply(1)
}

//KEYS *列出所有key，在wildcard包中给出了redis通配符的算法
// execKeys returns all keys matching the given pattern
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))      //取出*
	result := make([][]byte, 0)                              //存放所有符合要求的key
	db.data.ForEach(func(key string, val interface{}) bool { //遍历所有key返回符合的通配符
		if pattern.IsMatch(key) { //对于每一个key判断是否符合pattern
			result = append(result, []byte(key)) //把key(取出来是一个string转换为byte)追加到result切片中
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	//必须将上述指令注册到cmdTable中
	RegisterCommand("Del", execDel, -2)
	RegisterCommand("Exists", execExists, -2)
	RegisterCommand("Keys", execKeys, 2)
	RegisterCommand("FlushDB", execFlushDB, -1)  //FIUSHDB a b c
	RegisterCommand("Type", execType, 2)         //TYPE k1
	RegisterCommand("Rename", execRename, 3)     //RENAME k1 k2
	RegisterCommand("RenameNx", execRenameNx, 3) //KEYS *
}
