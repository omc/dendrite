package dendrite

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/fizx/logs"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type noOpReader struct{}
type rwStruct struct {
	io.Reader
	io.Writer
	io.Closer
}

type closeStruct struct {
	w *bufio.Writer
	c net.Conn
}

type libratoStruct struct {
	url       *url.URL
	responses chan string
	metrics   chan []byte
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
	case "librato":
		return NewLibratoReadWriter(u)
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

func NewLibratoReadWriter(u *url.URL) (io.ReadWriteCloser, error) {
	rw := new(libratoStruct)
	rw.url, _ = url.Parse(u.String())
	rw.url.Scheme = "https"
	rw.metrics = make(chan []byte, 1000)
	rw.responses = make(chan string, 1000)
	go rw.Loop()
	return rw, nil
}

func (rw *libratoStruct) Loop() {
	var msg []byte
	limit := 100
	msgs := make([][]byte, 0, limit)
	i := 0
	for {
		if i < limit {
			select {
			case msg = <-rw.metrics:
				msgs = append(msgs, msg)
				i += 1
				continue
			default:
			}
		}
		if len(msgs) > 0 {
			rw.Send(msgs)
			msgs = msgs[0:0]
			i = 0
		}
		time.Sleep(time.Second)
	}
}

func (rw *libratoStruct) Send(msgs [][]byte) {
	body := "{\"gauges\": [" + string(bytes.Join(msgs, []byte(","))) + "]}"
	logs.Info("Sending %d messages to librato: %s", len(msgs), rw.url.String())
	client := http.DefaultClient
	req, err := http.NewRequest("POST", "https://metrics-api.librato.com/v1/metrics", bytes.NewBufferString(body))
	if err != nil {
		logs.Error("error on http request: %s", err)
	} else {
		user := rw.url.User.Username()
		pass, _ := rw.url.User.Password()
		req.SetBasicAuth(user, pass)
		req.Header.Add("Content-type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			logs.Error("error on http response: %s", err)
		} else {
			logs.Info(resp.Status)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logs.Error("error reading http response: %s", err)
			} else {
				rw.responses <- resp.Status + "\n" + string(body)
			}
		}
	}
}

func (rw *libratoStruct) Close() error {
	return nil
}

func (rw *libratoStruct) Read(buf []byte) (int, error) {
	rsp := <-rw.responses
	n := copy(buf, rsp)
	if n < len(rsp) {
		return n, errors.New("response truncated")
	} else {
		return n, nil
	}
}

func (rw *libratoStruct) Write(msg []byte) (int, error) {
	rw.metrics <- msg
	return len(msg), nil
}
