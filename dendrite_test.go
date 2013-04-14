package main

import (
	"io"
	"os"
	"os/exec"
	"testing"
)

func _init() {
	os.RemoveAll("tmp")
	os.Mkdir("tmp", 0777)
}

func TestTcp(t *testing.T) {
	_init()
	cmd := exec.Command("./dendrite", "-o", "-d", "-f", "src/dendrite/data/conf.yaml")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Start()
	cmd.Wait()
}
