package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
