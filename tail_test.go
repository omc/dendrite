package dendrite

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
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
var output chan Record
var line = "{\"_file\":{\"Type\":0,\"Treatment\":0,\"Value\":\"solr.txt\"},\"_group\":{\"Type\":0,\"Treatment\":0,\"Value\":\"foo\"},\"_hostname\":{\"Type\":0,\"Treatment\":0,\"Value\":\"host.local\"},\"_offset\":{\"Type\":1,\"Treatment\":0,\"Value\":0},\"_time\":{\"Type\":3,\"Treatment\":0,\"Value\":1234567890},\"line\":{\"Type\":0,\"Treatment\":0,\"Value\":\"INFO: [1234567898765] webapp=/solr path=/select params={start=0&q=*:*&wt=ruby&fq=type:User&rows=30} hits=3186235 status=0 QTime=1\"}}"

func _tail_init() {
	StandardTimeProvider = new(FixedTimeProvider)
	output = make(chan Record, 100)
	offsetFile = path.Join(os.TempDir(), "test.txt")
	_ = os.Remove(offsetFile)
	parser = NewRegexpParser("host.local", "foo", "solr.txt", output, "(?P<line>.*[^\r])\r?\n", nil, 32768)
	tail = NewTail(parser, -1, "testdata/solr.txt", offsetFile, 0)
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
	tail = NewTail(parser, -1, "testdata/solr.txt", offsetFile, 0)
	if tail.Offset() != 747 {
		t.Errorf("initial offset was %d, not 747", tail.Offset())
	}
}

func TestReading(t *testing.T) {
	_tail_init()
	go tail.Poll()
	rec := <-output
	json, _ := json.Marshal(rec)
	if string(json) != line {
		t.Errorf("Oops, diff between\n    %s\n    %s", string(json), line)
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
