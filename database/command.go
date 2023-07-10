package database

import "strings"

//每一个command字段有一个执行方法，例如set/ping/等都是一个command结构体，结构体实现这个执行方法，在db中施加到db中
var cmdTable = make(map[string]*command) //记录系统中的所有指令，比如get/set等，这些指令和command之间的关系
//这个Map并发时只是只读的不用考虑并发安全

type command struct {
	executor ExecFunc
	arity    int // allow number of args, arity < 0 means len(args) >= -arity 参数的数量，比如ping需要几个参数
}

// RegisterCommand registers a new command
// arity means allowed number of cmdArgs, arity < 0 means len(args) >= -arity.
// for example: the arity of `get` is 2, `mget` is -2
//注册指令的方法，将新建的command放入到cmdTable中
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}
