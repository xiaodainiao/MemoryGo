package database

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"xiaodainiao/aof"
	"xiaodainiao/config"
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/logger"
	"xiaodainiao/resp/reply"
)

// StandaloneDatabase is a set of multiple database set
type StandaloneDatabase struct {
	dbSet      []*DB //一共16个DB,一个DB就是一个Dict,底层就是sync.Map
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase creates a redis database,
func NewStandaloneDatabase() *StandaloneDatabase {
	mdb := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	mdb.dbSet = make([]*DB, config.Properties.Databases)
	for i := range mdb.dbSet { //foe循环初始化所有DB
		singleDB := makeDB()
		singleDB.index = i
		mdb.dbSet[i] = singleDB
	}
	//初始化aofHandler
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(mdb)
		if err != nil {
			panic(err)
		}
		mdb.aofHandler = aofHandler

		//分DB有一个addaof方法，为了能够调用aofhandler(aof处理器)的aodAdd方法，写了一个匿名方法
		//改匿名方法调用了aof处理器的AddAof方法初始胡db的addaof
		for _, db := range mdb.dbSet {
			sdb := db
			sdb.addAof = func(line CmdLine) {
				mdb.aofHandler.AddAof(sdb.index, line)
			}
		}

	}

	return mdb
}

// Exec executes command
//调用DB的Exec方法，转交给对应的DB来处理
//set kv get k select 2
// parameter `cmdLine` contains command and its arguments, for example: "set key value"
func (mdb *StandaloneDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	//写一个recover，防止方法出现panic，防止整个进程崩溃
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0])) //把大写统一变为小写
	if cmdName == "select" {
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}
	// normal commands
	dbIndex := c.GetDBIndex()
	selectedDB := mdb.dbSet[dbIndex]
	return selectedDB.Exec(c, cmdLine)
}

// Close graceful shutdown database
func (mdb *StandaloneDatabase) Close() {

}

func (mdb *StandaloneDatabase) AfterClientClose(c resp.Connection) {
}

//用户发送来的指令修改DB
func execSelect(c resp.Connection, mdb *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0])) //字节转换为string在转换为int
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	//select 165232
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
