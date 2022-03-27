package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf *bufio.Writer
	dec *gob.Decoder
	enc *gob.Encoder
}

// 把conn的内容解码储存到h中
//
//实例化中dec:  gob.NewDecoder(conn)
func (c *GobCodec)ReadHeader(h *Header)error{
	return c.dec.Decode(h)
}

// 把conn的内容解码储存到body中
//
//实例化中dec:  gob.NewDecoder(conn)
func (c *GobCodec)ReadBody(body interface{})error{
	return c.dec.Decode(body)
}

func (c *GobCodec) Close() error{
	return c.conn.Close()
}

func (c *GobCodec) Write (h *Header,body interface{}) (err error) {
	defer func(){
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h);err != nil {
		log.Println("rpc codec: gob error encoding header:",err)
		return err
	}

	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec:gob error encoding body:",err)
	}

	return nil
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser)Codec{
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
