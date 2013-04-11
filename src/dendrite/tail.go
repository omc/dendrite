package dendrite

import (
	"bufio"
	"fmt"
	tails "github.com/ActiveState/tail"
	"io"

	"github.com/fizx/logs"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type TimeProvider interface {
	Now() time.Time
}

type SystemTimeProvider struct{}

func (*SystemTimeProvider) Now() time.Time {
	return time.Now()
}

var StandardTimeProvider TimeProvider = new(SystemTimeProvider)

type Tail struct {
	Path       string
	OffsetPath string
	Watcher    tails.FileWatcher
	Parser     Parser

	offset int64
	handle *os.File
}

func NewTail(parser Parser, path string, offsetPath string) *Tail {
	tail := new(Tail)
	tail.Path = path
	tail.OffsetPath = offsetPath
	tail.Parser = parser
	tail.Watcher = tails.NewInotifyFileWatcher(path)
	tail.LoadOffset()

	handle, err := os.Open(path)
	if err != nil {
		logs.Debug("Can't open file: ", path)
		return nil
	} else {
		tail.handle = handle
	}
	_, err = handle.Seek(tail.Offset(), 0)
	if err != nil {
		logs.Debug("Can't seek to %d in file: %s", tail.Offset(), path)
	}
	return tail
}

func (tail *Tail) Offset() int64 {
	return atomic.LoadInt64(&tail.offset)
}

func (tail *Tail) WriteOffset() {
	path := os.TempDir() + path.Base(tail.OffsetPath)
	temp, err := os.Create(path)
	if err != nil {
		logs.Debug("Can't create tempfile:", err)
	} else {
		_, err := fmt.Fprintf(temp, "%d\n", tail.Offset())
		if err != nil {
			logs.Debug("Can't write to tempfile:", err)
			temp.Close()
		} else {
			temp.Close()
			err := os.Rename(path, tail.OffsetPath)
			if err != nil {
				logs.Debug("Rename failed:", err)
			}
		}
	}
}

func (tail *Tail) LoadOffset() {
	file, err := os.Open(tail.OffsetPath)
	if err != nil {
		tail.WriteOffset()
	} else {
		reader := bufio.NewReader(file)
		str, err := reader.ReadString('\n')
		if err != nil {
			logs.Debug("Malformed offset file: ", err)
		} else {
			out, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
			if err != nil {
				logs.Debug("Malformed offset file: ", err)
			} else {
				logs.Debug("Found offset: %d", out)
				atomic.StoreInt64(&tail.offset, out)
			}
		}
		file.Close()
	}
}

func (tail *Tail) StartWatching() {
	go func() {
		c := tail.Watcher.ChangeEvents()
		for _, ok := <-c; ok; {
			tail.Poll()
		}
	}()
}

func (tail *Tail) Poll() {
	size := 16384
	buffer := make([]byte, size)
	for {
		len, err := tail.handle.Read(buffer)
		if err == io.EOF {
			return
		} else if err != nil {
			logs.Debug("read error: ", err)
			return
		} else {
			tail.Parser.Consume(buffer[0:len], &tail.offset)
			tail.WriteOffset()
		}
	}
}

func (tail *Tail) Close() {
	tail.handle.Close()
}
