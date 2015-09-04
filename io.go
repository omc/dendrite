package dendrite

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

type noOpReader struct{}
type rwStruct struct {
	io.Reader
	io.Writer
	io.Closer
}

type closeStruct struct {
	w *bufio.Writer
	c io.Closer
}

var EmptyReader = new(noOpReader)

func (er *noOpReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func NewReadWriter(u *url.URL) (io.ReadWriteCloser, error) {
	protocol := strings.Split(u.Scheme, "+")[0]
	switch protocol {
	case "file":
		realPath := u.Host + "/" + u.Path
		return NewFileReadWriter(strings.TrimRight(realPath, "/"))
	case "udp":
		return NewUDPReadWriter(u)
	case "tcp":
		return NewTCPReadWriter(u)
	case "tcps", "tcp+tls":
		panic("not implemented")
	case "http", "https":
		panic("not implemented")
	default:
		panic("unknown protocol")
	}
	return nil, nil //unreached
}

func NewFileReadWriter(path string) (io.ReadWriteCloser, error) {
	// fmt.Println(path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return nil, err
	}
	return &rwStruct{EmptyReader, file, file}, nil
}

func NewUDPReadWriter(u *url.URL) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("udp", u.Host)
	if err != nil {
		return nil, err
	}
	return &rwStruct{EmptyReader, conn, conn}, nil
}

func (cs *closeStruct) Close() error {
	cs.w.Flush()
	return cs.c.Close()
}

func NewTCPReadWriter(u *url.URL) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	return &rwStruct{r, w, &closeStruct{w, conn}}, nil
}
