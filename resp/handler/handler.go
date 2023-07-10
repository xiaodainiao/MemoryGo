package handler

import (
	"context"
	"xiaodainiao/cluster"
	"xiaodainiao/config"
	"xiaodainiao/database"
	databaseface "xiaodainiao/interface/database"
	"xiaodainiao/lib/logger"
	"xiaodainiao/lib/sync/atomic"
	"xiaodainiao/resp/connection"
	"xiaodainiao/resp/parser"
	"xiaodainiao/resp/reply"

	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

//处理RESP协议的内容
// RespHandler implements tcp.Handler and serves as a redis handler
type RespHandler struct {
	activeConn sync.Map              // *client -> placeholder
	db         databaseface.Database //redis核心
	closing    atomic.Boolean        // refusing new client and new request
}

//关闭其中的一个客户端连接
func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Close stops handler关闭整个handel,整个协议关掉说明redis关闭
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait  把每一个和客户端连接的都关闭
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}

// Handle receives and executes redis commands
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() { //判断是否在关闭中。如果是的话就干掉
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, struct{}{}) //将新建的client存在activeConn Map中

	ch := parser.ParseStream(conn) //把链接交给ParseStream,返回管道，将解析完的数据发送到管道，我只需要循环监听管道
	for payload := range ch {
		//error
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err协议错误，用户发送的数据不是规定的格式
			errReply := reply.MakeErrReply(payload.Err.Error()) //前面拼上-后面拼上\r\n,写会到客户端
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		//exec正常逻辑
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}

		//payload.Data返回是一维数组，我需要把它转换为二维数组
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// MakeHandler creates a RespHandler instance
//resp协议层持有database核心，他把解析好的指令传给我们
func MakeHandler() *RespHandler {
	var db databaseface.Database
	//db = database.NewEchoDatabase()
	//首先判断配置文件中设置了集群模式吗，如果有的话调用集群的database如果没有就调用单击dtabase
	//其中集群的database调用了单机版的database
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}

	return &RespHandler{
		db: db,
	}
}
