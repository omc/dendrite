package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func bash(str string) {
	run("bash", "-c", str)
}

func run(str ...string) {
	cmd := exec.Command(str[0], str[1:]...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Start()
	cmd.Wait()
}

func _init(t *testing.T) {
	os.RemoveAll("tmp")
	os.Mkdir("tmp", 0777)
	matches, _ := filepath.Glob("src/dendrite/data/*")
	for _, m := range matches {
		os.Chtimes(m, time.Now(), time.Now())
	}
}

func TestTruncation(t *testing.T) {
	_init(t)
	bash("cp src/dendrite/data/solr.txt tmp/solr.txt")
	go func() {
		time.Sleep(time.Second / 2)
		bash("cat /dev/null > tmp/solr.txt")
		bash("cat src/dendrite/data/solr.txt >> tmp/solr.txt")
	}()
	run("./dendrite", "-q", "1", "-d", "-f", "src/dendrite/data/truncate.yaml")
	bytes, err := ioutil.ReadFile("tmp/out.json")
	if err != nil {
		t.Error(err)
	}
	str := string(bytes)
	arr := strings.Split(strings.TrimSpace(str), "\n")
	if len(arr) != 2000 {
		t.Error(len(arr), "not 2000")
	}
}

func TestBackfill(t *testing.T) {
	_init(t)
	bash("cp src/dendrite/data/solr.txt tmp/solr.txt")
	run("./dendrite", "-q", "0", "-d", "-f", "src/dendrite/data/backfill.yaml")
	bytes, err := ioutil.ReadFile("tmp/out.json")
	if err != nil {
		t.Error(err)
	}
	str := string(bytes)
	arr := strings.Split(strings.TrimSpace(str), "\n")
	if len(arr) != (600 / 130) {
		t.Error(len(arr), "not", (600 / 130))
	}
}

func TestTcp(t *testing.T) {
	_init(t)
	run("./dendrite", "-q", "0", "-d", "-f", "src/dendrite/data/conf.yaml")
	bytes, err := ioutil.ReadFile("tmp/out.json")
	if err != nil {
		t.Error(err)
	}
	str := string(bytes)
	arr := strings.Split(strings.TrimSpace(str), "\n")
	if len(arr) != 1100 {
		t.Error(len(arr), "not 1100")
	}
	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(arr[0]), &m)
	if err != nil {
		t.Error(err)
	}
	if m["qtime"] != 1.0 {
		t.Error(m["qtime"], "wasn't a numeric 1 ")
	}
}

func TestCookbooks(t *testing.T) {
	sentinel := "# -- log line --\n"
	rglob := regexp.MustCompile("(\n[ ]+glob:).*")
	rlog := regexp.MustCompile("# (.*)")
	rout := regexp.MustCompile("(?s)# -- output --.*?# (.+?# })")
	matches, _ := filepath.Glob("cookbook/*.yaml")
	for _, m := range matches {
		bytes, err := ioutil.ReadFile(m)
		if err != nil {
			t.Fatal("can't open", m)
		}

		strs := strings.Split(string(bytes), sentinel)
		for i, str := range strs {
			if i == 0 {
				continue
			}

			os.RemoveAll("tmp")
			os.Mkdir("tmp", 0777)
			os.Mkdir("tmp/conf.d", 0777)
			exec.Command("cp", "src/dendrite/data/conf.yaml", "tmp").Run()

			log := rlog.FindStringSubmatch(str)[1]
			out := rout.FindStringSubmatch(str)[1]
			out = strings.Replace(out, "\n#", " ", -1)

			ioutil.WriteFile("tmp/foo.log", []byte(log+"\n"), 0777)

			bytes = rglob.ReplaceAll(bytes, []byte("$1 tmp/foo.log"))
			ioutil.WriteFile("tmp/conf.d/sub.yaml", bytes, 0777)
			run("./dendrite", "-q", "0", "-d", "-f", "tmp/conf.yaml", "-l", "tmp/test.log")

			var expected map[string]interface{}
			var actual map[string]interface{}
			json.Unmarshal([]byte(out), &expected)
			actualBytes, _ := ioutil.ReadFile("tmp/out.json")
			json.Unmarshal(actualBytes, &actual)
			if len(expected) == 0 {
				t.Error("malformatted expect")
			}
			for k, _ := range expected {
				if fmt.Sprintf("%s", actual[k]) != fmt.Sprintf("%s", expected[k]) {
					t.Fatal("mismatch on", k, "[", actual[k], "]", "[", expected[k], "]")
				}
			}
		}
	}
}
