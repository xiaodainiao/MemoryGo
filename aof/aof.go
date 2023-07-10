package aof

import (
	"io"
	"os"
	"strconv"
	"xiaodainiao/config"
	databaseface "xiaodainiao/interface/database"
	"xiaodainiao/lib/logger"
	"xiaodainiao/lib/utils"
	"xiaodainiao/resp/connection"
	"xiaodainiao/resp/parser"
	"xiaodainiao/resp/reply"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16 //65535
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler receive msgs from channel and write to AOF file
type AofHandler struct {
	db          databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int //记录上一条指令写在哪一个DB上，看看是不是需要切换
}

// NewAOFHandler creates a new aof.AofHandler
//初始化上面结构体
func NewAofHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	handler.LoadAof() //系统刚启动的时候，需要把写在硬盘里的aof文件加载上来（恢复）
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	//aofFile, err := os.OpenFile("appendonly.aof", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile

	//数据库引擎在执行的时候把他要写的aof的条目赛到这个channel里面，不断的从chan中取出数据落在磁盘上
	handler.aofChan = make(chan *payload, aofQueueSize)
	go func() { //handleaof要跑到一个协程了，源源不断从chan中获取
		handler.handleAof()
	}()
	return handler, nil
}

/*
把写到磁盘文件的操作记录塞到channnel中，它不落盘（赛道channel中不落盘的原因就是redis操作时要等待落盘太慢了）
//有另一个协程给它落盘
AddAof send command to aof goroutine through channel
payload(set k v) -> aofChan
每进行操作就会调用Addof把具体的指令塞到这个channel里面去
*/
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) { //CmdLine指令
	if config.Properties.AppendOnly && handler.aofChan != nil { //判断是不是开启了aof功能
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

//从channel中不断取
// handleAof listen aof channel and write into file
//payload(set k v ) <- aofChan(落盘)
//handleaof要跑到一个协程了，源源不断从chan中获取
func (handler *AofHandler) handleAof() {
	// serialized execution
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { //DB号不一样插入select语句，//每次循环查看我的dbIndex和之前的currentDB是不是同一个DB
			// select db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

// LoadAof read aof file
//写的时候就是按照RESP往磁盘中写的，当作把用户通过TCP发送的指令，重新回复一遍
func (handler *AofHandler) LoadAof() {

	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Warn(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)
	fakeConn := &connection.Connection{} // only used for save dbIndex构造一个空的，仅仅记录selectDB
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil { //空指令
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply) //aof文件中都是二维字节数组[][]byte
		if !ok {                                //如果遇到+ok\r\n就要丢弃
			logger.Error("require multi bulk reply")
			continue
		}

		ret := handler.db.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) { //判断是否是错误回复
			logger.Error("exec err", err)
		}
	}
}
