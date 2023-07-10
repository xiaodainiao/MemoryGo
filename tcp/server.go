package tcp

import (
	"context"
	"xiaodainiao/interface/tcp"
	"xiaodainiao/lib/logger"

	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServerWithSignal(ctx *Config, handler tcp.Handler) error {

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	//当OS发送任何信号，通过Notify你都传给我这个sigChan
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", ctx.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServer(listener, handler, closeChan)
	return nil

}

func ListenAndServer(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	/*
		如果用户直接关闭窗口，此时还没走到defer，导致连接没关闭，此时需要传进来的closeChan起作用
		上面调用方会感知系统的信号，当系统给进程发送关闭信号时，此时就会通过channel通知ListenAndserver的channel
		通知到了就关闭listener和handler
	*/

	/*当退出的时候关闭listener和handeler并且收到信号时关闭listener和handler*/
	go func() {
		<-closeChan
		logger.Info("shut down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close() //程序退出时关闭listener和业务的handler
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup //等待所有的连接退出，如果有一个socket连接错误，不要立刻退出，因为之前的连接可能在处理业务

	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		waitDone.Add(1) //给等待队列增加一个
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
