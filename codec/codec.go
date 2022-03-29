package codec

import "io"

type Header struct {
	ServiceMethod string //服务名和方法名，通常与 Go 语言中的结构体和方法相映射
	Seq uint64  //请求的序号，也可以认为是某个请求的 ID，用来区分不同的请求
	Error string  //错误信息，客户端置为空，服务端如果如果发生错误，将错误信息置于 Error 中
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error  // 把conn的内容解码储存到Header中
	ReadBody(interface{}) error  // 把conn的内容解码储存到interface{}中
	Write(*Header,interface{}) error  // 把Header和interface{}的内容写到conn中（就是相当于把内容发送给客户端）
}

type NewCodecFunc func(io.ReadWriteCloser)Codec

type Type string

const (
	GobType Type = "application/gob"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
