package database

import (
	"xiaodainiao/interface/database"
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/utils"
	"xiaodainiao/resp/reply"
)

//实现GET SET SETNX GETSET STRLEN

//func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
//	entity, ok := db.GetEntity(key)
//	if !ok {
//		return nil, nil
//	}
//	bytes, ok := entity.Data.([]byte)
//	if !ok {
//		return nil, &reply.WrongTypeErrReply{}
//	}
//	return bytes, nil
//}

// execGet returns string value bound to the given key
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key) //返回一个DateEntity,里面是一个Date interface{}
	if !exists {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes)

}

//这里只实现SET k v
// execSet sets string value and time to live to the given key
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)
	db.addAof(utils.ToCmdLine2("set", args...))
	//刷盘由于addAof方法在aof.go文件中，而aofadd属于aofhandler的，而执行set k v是在db执行（db并没有aofhandler,aofhandler在database这是fb的上一层）所以调用不到这个方法，
	return &reply.OkReply{}
}

// execSetNX sets string if not exists
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	result := db.PutIfAbsent(key, entity) //不存在的时候才操作
	db.addAof(utils.ToCmdLine2("setnx", args...))
	return reply.MakeIntReply(int64(result)) //上一个set回复ok这个回复0或者1
}

//GetSet k1 v1首先获取k1的值，然后将k1设置为新的值，返回k1原来的值
// execGetSet sets value of a string-type key and returns its old value
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]

	entity, exists := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{Data: value})
	db.addAof(utils.ToCmdLine2("getset", args...))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	old := entity.Data.([]byte) //转到[]byte数组

	return reply.MakeBulkReply(old)
}

//STRLEN  K -> 'value'获取key的value长度
// execStrLen returns len of string value bound to the given key
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	old := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(old)))
}

func init() {
	RegisterCommand("Get", execGet, 2)  //get K1
	RegisterCommand("Set", execSet, -3) //set k1 v1
	RegisterCommand("SetNx", execSetNX, 3)
	RegisterCommand("GetSet", execGetSet, 3)
	RegisterCommand("StrLen", execStrLen, 2)
}
