package dendrite

import (
	"bufio"
	"io"
	"net"
	"net/url"
	"os"
)

type noOpReader struct{}
type rwStruct struct {
	io.Reader
	io.Writer
}

var EmptyReader = new(noOpReader)

func (er *noOpReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func NewReadWriter(u *url.URL) (io.ReadWriter, error) {
	switch u.Scheme {
	case "file":
		return NewFileReadWriter(u.Path)
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

func NewFileReadWriter(path string) (io.ReadWriter, error) {
	file, err := os.OpenFile(path, 0, 0777)
	if err != nil {
		return nil, err
	}
	return &rwStruct{EmptyReader, file}, nil
}

func NewUDPReadWriter(u *url.URL) (io.ReadWriter, error) {
	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	return &rwStruct{EmptyReader, conn}, nil
}

func NewTCPReadWriter(u *url.URL) (io.ReadWriter, error) {
	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	return &rwStruct{bufio.NewReader(conn), bufio.NewWriter(conn)}, nil
}
