package codec

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestNewGobCodec(t *testing.T) {
	var a io.Writer
	a = os.Stdout
	b := bufio.NewWriter(a)
	_,_ = b.Write([]byte("hello"))
	fmt.Println("111")
	b.Flush()
	fmt.Println("222")

}
