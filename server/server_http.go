package server

import (
	"GeekRPC"
	"io"
	"log"
	"net/http"
)

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type","text/plain;charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_,_ = io.WriteString(w,"405 must CONNECT\n")
		return
	}

	conn,_,err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ",req.RemoteAddr,": ",err.Error())
		return
	}
	_,_ = io.WriteString(conn,"HTTP/1.0"+GeekRPC.Connected+"\n\n")
	server.ServeConn(conn)
}

func (server *Server) HandleHTTP() {
	http.Handle(GeekRPC.DefaultRPCPath,server)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
