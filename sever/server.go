package sever

import (
	"GeekRPC/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const  MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int  // MagicNumber marks this's a geerpc request
	CodecType codec.Type  // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// Server represents an RPC Server.
type Server struct {}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

type request struct {
	h *codec.Header
	argv reflect.Value
	replyv reflect.Value
}

//读取cc，解码后返回
func (server *Server) readRequestHeader(cc codec.Codec)(*codec.Header,error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error: ",err)
		}
		return nil,err
	}
	return &h,nil
}

// 把cc中的信息解码后以request的形式返回
func (server *Server) readRequest(cc codec.Codec)(*request,error) {
	h,err := server.readRequestHeader(cc)
	if err != nil {
		return nil,err
	}

	req := &request{h:h}
	// TODO: now we don't know the type of request argv
	// day 1, just suppose it's string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err: ",err)
	}

	return req,nil
}

//把h和body的信息写进cc中
func (server *Server) sendResponse(cc codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex){
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h,body); err != nil {
		log.Println("rpc server: write response error: ",err)
	}
}

//把req的信息加工添加一些内容后写进cc中
func (server *Server) handleRequest(cc codec.Codec,req *request,sending *sync.Mutex,wg *sync.WaitGroup){
	// TODO, should call registered rpc methods to get the right replyv
	// day 1, just print argv and send a hello message
	defer wg.Done()
	log.Println(req.h,req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d",req.h.Seq))
	server.sendResponse(cc,req.h,req.replyv.Interface(),sending)

}

var invalidRequest = struct{}{}

//读取cc中的内容，
//
//加工添加一些额外信息进去后再写进cc中
func (server *Server) serveCodec(cc codec.Codec) {
	// Todo
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req,err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc,req.h,invalidRequest,sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc,req,sending,wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
//
//读取conn，解码成option形式
//
//通过option检查magicNumber是否符合，以及确认编码方法CodecFunc
//
//通过CodecFunc对conn进行编码，传进serveCodec中，实现对信息的加工和返回
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func(){
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt);err!=nil{
		log.Println("rpc server: options error: ",err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x ",opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec Type %s ",opt.CodecType)
		return
	}

	server.serveCodec(f(conn))


}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
//接收net.Listener返回的lis，通过lis.Accept()返回的conn,实现conn的编码解码以及处理信息后再写进conn中
func (server *Server) Accept (lis net.Listener) {
	for {
		conn,err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:",err)
			return
		}
		go server.ServeConn(conn)
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}