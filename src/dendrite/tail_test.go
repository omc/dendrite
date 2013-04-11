package dendrite

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type FixedTimeProvider struct{}

func (*FixedTimeProvider) Now() time.Time {
	return time.Unix(1234567890, 0)

}

var parser Parser
var offsetFile string
var tail *Tail
var output chan string
var line = "{\"_file\":\"solr.txt\",\"_group\":\"foo\",\"_offset\":0,\"_time\":1234567890,\"line\":\"INFO: [1234567898765] webapp=/solr path=/select params={start=0&q=*:*&wt=ruby&fq=type:User&rows=30} hits=3186235 status=0 QTime=1\"}"

func _tail_init() {
	StandardTimeProvider = new(FixedTimeProvider)
	output = make(chan string, 100)
	offsetFile = os.TempDir() + "test.txt"
	_ = os.Remove(offsetFile)
	encoder := new(JsonEncoder)
	parser = NewRegexpParser("foo", "solr.txt", output, "(?P<line>.+)\n", nil, encoder)
	tail = NewTail(parser, "data/solr.txt", offsetFile)
}

func TestStartsAtZero(t *testing.T) {
	_tail_init()
	if tail.Offset() != 0 {
		t.Error("initial offset wasn't zero")
	}
}

func TestStartsAtOffset(t *testing.T) {
	_tail_init()
	ioutil.WriteFile(offsetFile, []byte("747\n"), 0777)
	tail = NewTail(parser, "data/solr.txt", offsetFile)
	if tail.Offset() != 747 {
		t.Errorf("initial offset was %d, not 747", tail.Offset())
	}
}

func TestReading(t *testing.T) {
	_tail_init()
	go tail.Poll()
	str := <-output
	if str != line {
		t.Error("Oops, got ", str)
	}
}

func TestOffsetUpdated(t *testing.T) {
	_tail_init()
	go tail.Poll()
	_ = <-output
	if tail.Offset() == 0 {
		t.Error("offset was zero")
	}
}
