package database

import (
	"strings"
	"xiaodainiao/datastruct/dict"
	"xiaodainiao/interface/database"
	"xiaodainiao/interface/resp"
	"xiaodainiao/resp/reply"
)

//dict是redis最底层的数据结构，它的上一层是db,Redis默认16个db
type DB struct {
	index int
	// key -> DataEntity
	data   dict.Dict
	addAof func(CmdLine) //给DB一个方法，后面吧AddAof赋值给这个方法，这样就可以调用Aof这个方法了
}

//set k v
//执行函数（函数声明）：所有redis指令，比如ping,get,set,setnx给他们一个实现
type ExecFunc func(db *DB, args [][]byte) resp.Reply
type CmdLine = [][]byte

func makeDB() *DB {
	db := &DB{
		data:   dict.MakeSyncDict(),
		addAof: func(line CmdLine) {}, //调用aofhandler->loadaof(把之前的指令回放一遍，此时执行set的时候又会调用db.addaof方法)
		//刚启动时调用的addAof func(line CmdLine)没有初始化会出问题，空方法New就出问题
	}
	return db
}

//内核执行完一个指令后set k v要返回执行的结果，
/*
1. 首先判断用户发送的什么指令是PING还是SET SETNX,这个指令是第一个切片
*/
func (db *DB) Exec(c resp.Connection, cmdLine [][]byte) resp.Reply {

	cmdName := strings.ToLower(string(cmdLine[0])) //拿到指令名称，作为string放入var cmdTable = make(map[string]*command)
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	//SET K
	if !validateArity(cmd.arity, cmdLine) { //校验参数个数是否合法
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.executor
	//SET K V -> K V
	return fun(db, cmdLine[1:])
}

//校验参数格式是否合法
//SET K V -> arity = 3
//EXISTS k1 k2 k3 k4 -> artiy = -2 符号只是标记变长还是定长（负的最小值）
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 { //说明是定长的
		return argNum == arity
	}
	return argNum >= -arity //变长的
}

/* ----------------------- data Access ---------------- */

// GetEntity returns DataEntity bind to given key
//用key获得value
//DataEntity内部是空接口指代redis所有数据类型
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {

	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity) //将raw强制转换为DataEntity，因为Get返回的是空接口，而DataEntity是结构体
	return entity, true
}

// PutEntity a DataEntity into DB
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists edit an existing DataEntity
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)

}

// PutIfAbsent insert an DataEntity only if the key not exists
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove the given key from db
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Removes the given keys from db
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys { //遍历所有入参
		_, exists := db.data.Get(key) //判断是否存在
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Flush clean database
func (db *DB) Flush() {
	db.data.Clear()

}
