package parser

import (
	"bufio"
	"errors"
	"xiaodainiao/interface/resp"
	"xiaodainiao/lib/logger"
	"xiaodainiao/resp/reply"

	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

/*信号解析器：用户发过来的有信号，\r\n，$等请求，解析为自己识别的内容
客户端发送的数据和服务器回复的数据都是相同的格式（5种通信格式）*/

type Payload struct {
	Data resp.Reply //客户端发送的数据
	Err  error
}

type readState struct {
	readingMultiLine  bool     //redis解析一行还是多行数据
	expectedArgsCount int      //现在读取的命令有几个参数例如set key value这是三个参数（目标参数数量）
	msgType           byte     //用户发送的消息类型是数组还是什么
	args              [][]byte //已经解析的参数
	bulkLen           int64    //数据块的长度（整个字节的长度）$3就代表后面有3个字符长度
}

//判断解析器是否解析完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

/*提前做一个异步的解析，把redis核心业务处理和协议解析并发跑
这样发送过来一个语句就执行，可以并发解析下一条*/
// ParseStream reads data from io.Reader and send payloads through channel
//大写说明是对外提供的接口，TCP层可以调用该接口
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch //TCP调用后，返回一个chan
}

//TCP层做了一个io.reader读取客户端传给我们的字节流
//parse0会源源不断的把数据塞进管道
func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() { //防止for循环中有异常退出来
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for { //不断循环解析，直到断开，就跳出循环
		// read line
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // encounter io err, stop read  IO错误
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			// protocol err, reset read state          协议错误，不需要结束，只需要告诉用户错误了
			ch <- &Payload{
				Err: err,
			}
			state = readState{} //把状态清空
			continue
		}

		// parse line 判断是否是多行解析模式，根据是不是多行解析模式来改变解析行为
		//解析器初始化状态，可能遇到*、$、-、+、:,就可以根据这些，将解析器初始化为单行或者多行模式
		// *3\r\n$3\r\nSET\r\n$3\r\nKEy\r\n$5\r\nvalue\r\n
		/*
			1. 如果用户发送SET Key Value
			2. 解析器遇到*3,此时进入该if判断，判断是否是多行模式，如果是的话就改为多行模式，然后记录下来我要接收三个字符串
			3. 此时continue,又读了一行，把$3\r\n读进来，此时进入不了if判断了，只能进入else,因为它已经改为多行模式
		*/

		/*
			1. 你的上层调用ParseStream，调用ParseStream之后返回一个channel(不是同步阻塞的)
			2. 业务层（redis核心）一直监听channel,看有没有新数据产生，如果有新的数据就是用户发送过来的指令
			3. parse0协程是一个用户一个，解析器为每一个用户生成一个解析器
			4. 用户发送来的指令，都由属于该用户的解析器来解析
		*/

		/*
			1. 写一个handler来处理用户请求，把该请求转发给解析器
		*/

		if !state.readingMultiLine {
			// receive new response
			if msg[0] == '*' {
				// multi bulk reply多行解析
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{}, //不是给客户返回，是给redis内核返回
					}
					state = readState{} // reset state
					continue
				}
			} else if msg[0] == '$' { // bulk reply
				err = parseBulkHeader(msg, &state) //$4\r\nPING\r\n多行
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.bulkLen == -1 { // null bulk reply  $-1\r\n 空指令
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else { //:或者-或者+
				// single line reply
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{} // reset state
				continue
			}
		} else {
			// receive following bulk reply
			err = readBody(msg, &state) //此时会判断多行模式下，读进来的是$还是数字
			//$3SET和$5Value多会进入readbody

			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// if sending finished
			if state.finished() { //判断用户发送的数据是否已经完全被接收
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// *3\r\n$3\r\nSET\r\n$3\r\nKEy\r\n$5\r\nvalue\r\n
/*
分为俩种情况
1. 无$\r\n切分
2. 之前读到了$字符，就必须按照它规定的字符个数进行切分*/
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { //1. \r\n切分
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err //第二个参数true代表有io错误，第三个参数就是真正错误
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { //第二种切分方法
		msg = make([]byte, state.bulkLen+2)  //例如$3\r\nSET\r\n,实际读5个把\r\n读进来
		_, err = io.ReadFull(bufReader, msg) //把bufReader内容，强行塞满msg
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

/*上面redline只是通过\r\n切割但是具体含义不知道，我们需要把parese状态改变
相当于把*3中的3读进来，改变状态*/
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64                                                  //例如*3读到*号就把3取出来
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32) //输入一个字符，把字符代表的数字取出来
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 { //*3\r\n$3\r\nSET\r\n$3\r\nKEy\r\n$5\r\nvalue\r\n
		// first line of multi bulk reply   将3读进来了，还要继续读后面的3个参数
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

/*
$=================================$4\r\nPING\r\n========================================
*/
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

//=======================+OK\r\n  -err\r\n :5\r\n  单行就是这三种=======================
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n") //把后缀\r\n切掉
	var result resp.Reply                          //该接口代表客户端和服务端的双向回复内容，因为格式都一样
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:]) //获的OK
	case '-': // err reply
		result = reply.MakeErrReply(str[1:]) //获得ERR
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64) //把:取出来，按照10进制解析为int64
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// read the non-first lines of multi bulk reply or bulk reply
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
