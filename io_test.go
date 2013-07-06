package dendrite

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewFileReadWriter(t *testing.T) {
	if err := os.RemoveAll("tmp.out"); err != nil {
		t.Fatal(err)
	}
	hai := []byte("hai world")
	u, err := url.Parse("file+json://tmp.out")
	check(t, err)

	rw, err := NewReadWriter(u)
	check(t, err)

	n, err := rw.Write(hai)
	check(t, err)
	if n != 9 {
		t.Fatal("wrote wrong bytes")
	}

	rw.Close()

	out, err := ioutil.ReadFile("tmp.out")
	check(t, err)

	if bytes.Compare(out, hai) != 0 {
		t.Fatal("didn't read back bytes", out, hai)
	}
}

func TestNewUDPReadWriter(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4009")
	check(t, err)

	conn, err := net.ListenUDP("udp", addr)

	hai := []byte("hai world")
	u, err := url.Parse("udp+json://localhost:4009")
	check(t, err)

	rw, err := NewReadWriter(u)
	check(t, err)

	n, err := rw.Write(hai)
	check(t, err)
	if n != 9 {
		t.Fatal("wrote wrong bytes")
	}

	out := make([]byte, 100)

	n, addr, err = conn.ReadFromUDP(out)
	if n != 9 {
		t.Fatal("read wrong bytes", n)
	}

	if bytes.Compare(out[0:n], hai) != 0 {
		t.Fatal("didn't read back bytes", out, hai)
	}
}

func connHandler(t *testing.T, ln net.Listener, ch chan []byte) {
	conn, err := ln.Accept()
	check(t, err)
	buf := make([]byte, 100)

	n, err := conn.Read(buf)
	check(t, err)
	ch <- buf[0:n]
}

func TestNewTCPReadWriter(t *testing.T) {
	ch := make(chan []byte, 1)
	conn, err := net.Listen("tcp", "127.0.0.1:4009")
	go connHandler(t, conn, ch)

	hai := []byte("hai world")
	u, err := url.Parse("tcp+json://localhost:4009")
	check(t, err)

	rw, err := NewReadWriter(u)
	check(t, err)

	n, err := rw.Write(hai)

	check(t, err)
	if n != 9 {
		t.Fatal("wrote wrong bytes")
	}
	rw.Close()

	out := <-ch

	if bytes.Compare(out[0:n], hai) != 0 {
		t.Fatal("didn't read back bytes", out, hai)
	}
}
