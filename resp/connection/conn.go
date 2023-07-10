package connection

import (
	"net"
	"sync"
	"time"
	"xiaodainiao/lib/sync/wait"
)

//该结构体代表协议层对每一个连接上的客户端的描述
type Connection struct {
	conn net.Conn
	// waiting until reply finished
	waitingReply wait.Wait
	// lock while handler sending response
	mu sync.Mutex //在操作一个客户时要上锁
	// selected db
	selectedDB int //redis有16个隔离的DB
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

//实现Connection interface
//使用TCP连接发送给客户端
func (c *Connection) Write(b []byte) error {
	//直接使用c.conn.Write(b)就完事，但是有一些特殊情况
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock() //同一时刻只能有一个协程，给客户端回写数据
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex returns selected db
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB selects a database
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}

// Close disconnect with the client
func (c *Connection) Close() error { //关闭时不能立刻关闭，要等待这次通信结束后才能关闭
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
