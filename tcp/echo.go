package tcp

import (
	"bufio"
	"context"
	"xiaodainiao/lib/logger"
	"xiaodainiao/lib/sync/atomic"
	"xiaodainiao/lib/sync/wait"

	"io"
	"net"
	"sync"
	"time"
)

/*业务处理：记录所有的客户端信息，并且你发什么我回复什么*/

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (e *EchoClient) Close() error { //关闭统一的接口
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandle struct {
	activeConn sync.Map //记录有多少个客户连接
	closing    atomic.Boolean
}

//这个handler是TCP的handler要改成RESP的handler
func MakeHandler() *EchoHandle {
	return &EchoHandle{} //做一个初始化（因为Map和close都不需要特殊处理）
}

func (handler *EchoHandle) Handle(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() { //判断一下状态是否在关闭
		_ = conn.Close()
	}
	client := &EchoClient{ //如果是真的连接，则记录连接客户端信息
		Conn: conn,
	}
	handler.activeConn.Store(client, struct{}{}) //只需要键不需要值，因此value传入空结构体
	reader := bufio.NewReader(conn)              //将conn的数据读进buf中
	for {
		//一换行就把你发的消息写回去
		msg, err := reader.ReadString('\n') //这个函数会尝试从 reader 中读取一行数据，直到遇到换行符 ('\n') 为止。一旦读取到一行数据，它将作为返回值被返回，并在下一次循环开始时再次执行该语句。
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1) //业务没做完之前不要关闭，除非等待10s还没做完
		b := []byte(msg)      //将msg转换为字节在写回去
		_, _ = conn.Write(b)  //Write需要一个字节流而msg是一个字符串
		client.Waiting.Done()
	}
}

//整个业务都要关了，所有记录在map中的客户端都干掉
func (handler *EchoHandle) Close() error {
	logger.Info("handler shutting down")
	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value interface{}) bool { //对每一个key value做关闭操作
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		return true //只有return true才会将你的这个关闭施加到下一个key,v上
	})
	return nil
}
