package client

import (
	"GeekRPC/codec_day1"
	"GeekRPC/codec_day1/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Call struct {
	Seq uint64
	ServiceMethod string
	Args interface{}
	Reply interface{}
	Error error
	Done chan *Call
}

func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	cc codec.Codec
	opt *codec_day1.Option
	sending sync.Mutex
	header codec.Header
	mu sync.Mutex
	seq uint64
	pending map[uint64]*Call //存储未处理完的请求，键是编号，值是 Call 实例
	closing bool //closing 是用户主动关闭的
	shutdown bool //shutdown 置为 true 一般是有错误发生
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing {
		return ErrShutdown
	}

	client.closing = true

	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

//将参数 call 添加到 client.pending 中，并更新 client.seq
func (client *Client) registerCall(call *Call)(uint64,error)  {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing || client.shutdown {
		return 0,ErrShutdown
	}

	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq,nil

}

//根据 seq，从 client.pending 中移除对应的 call，并返回
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending,seq)
	return call
}

//服务端或客户端发生错误时调用，将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true

	for _,call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}

		call := client.removeCall(h.Seq)
		switch{
		case call == nil:
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body" + err.Error())
			}
			call.done()
		}

	}
	client.terminateCalls(err)
}

func NewClient(conn net.Conn,opt *codec_day1.Option) (*Client,error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s",opt.CodecType)
		log.Println("rpc client: codec error:",err)
		return nil,err
	}

	if err := json.NewEncoder(conn).Encode(opt);err != nil {
		log.Println("rpc client: options error: ",err)
		_ = conn.Close()
		return nil,err
	}

	return newClientCodec(f(conn),opt),nil
}

func newClientCodec(cc codec.Codec,opt *codec_day1.Option) *Client {
	client := &Client {
		seq: 1,
		cc: cc,
		opt: opt,
		pending: make(map[uint64]*Call),
	}

	go client.receive()

	return client
}

func parseOption(opts ...*codec_day1.Option) (*codec_day1.Option,error){
	if len(opts) == 0 || opts[0] == nil {
		return codec_day1.DefaultOption,nil
	}

	if len(opts) != 1{
		return nil,errors.New("number of options is more than 1")
	}

	opt := opts[0]
	opt.MagicNumber = codec_day1.DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = codec_day1.DefaultOption.CodecType
	}

	return opt,nil
}

func Dial(network,address string,opts ...*codec_day1.Option)(client *Client,err error){
	opt,err := parseOption(opts...)
	if err != nil {
		return nil,err
	}

	conn,err := net.Dial(network,address)
	if err != nil {
		return nil,err
	}

	defer func(){
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn,opt)
}

func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	seq,err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	if err := client.cc.Write(&client.header,call.Args);err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) Go(serviceMethod string,args,reply interface{},done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call,10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args: args,
		Reply: reply,
		Done: done,
	}
	client.send(call)
	return call
}

func (client *Client) Call(serviceMethod string,args,reply interface{}) error {
	call := <- client.Go(serviceMethod,args,reply,make(chan *Call,1)).Done
	return call.Error
}