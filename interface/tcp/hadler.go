package tcp

import (
	"context"
	"net"
)

//写个接口，代表抽象的业务逻辑，也就是TCP服务只处理TCP这一层的连接（只是处理连接）
//具体的业务扔给handler做
type Handler interface { //业务逻辑的处理接口
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
