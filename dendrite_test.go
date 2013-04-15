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

func _init(t *testing.T) {
	os.RemoveAll("tmp")
	os.Mkdir("tmp", 0777)
	matches, _ := filepath.Glob("src/dendrite/data/*")
	for _, m := range matches {
		os.Chtimes(m, time.Now(), time.Now())
	}
}

func TestTcp(t *testing.T) {
	_init(t)
	cmd := exec.Command("./dendrite", "-o", "-d", "-f", "src/dendrite/data/conf.yaml")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Start()
	cmd.Wait()
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
	rglob := regexp.MustCompile("(\n[ ]+glob:).*")
	rlog := regexp.MustCompile("# -- log line --\\s*# (.*)")
	rout := regexp.MustCompile("(?s)# -- output --\\s*# (.+?)\n[^#]")
	matches, _ := filepath.Glob("cookbook/*.yaml")
	for _, m := range matches {
		os.RemoveAll("tmp")
		os.Mkdir("tmp", 0777)
		os.Mkdir("tmp/conf.d", 0777)
		exec.Command("cp", "src/dendrite/data/conf.yaml", "tmp").Run()

		bytes, err := ioutil.ReadFile(m)
		if err != nil {
			t.Fatal("can't open", m)
		}

		log := rlog.FindStringSubmatch(string(bytes))[1]
		fmt.Println(string(log))
		out := rout.FindStringSubmatch(string(bytes))[1]
		out = strings.Replace(out, "\n#", " ", -1)

		ioutil.WriteFile("tmp/foo.log", []byte(log+"\n"), 0777)

		bytes = rglob.ReplaceAll(bytes, []byte("$1 tmp/foo.log"))
		ioutil.WriteFile("tmp/conf.d/sub.yaml", bytes, 0777)
		cmd := exec.Command("./dendrite", "-o", "-d", "-f", "tmp/conf.yaml")
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		go io.Copy(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
		cmd.Start()
		cmd.Wait()

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
				t.Error("mismatch on", k, actual[k], expected[k])
			}
		}
	}
}
