package process

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	p := NewProcess(exec.Command("sleep", "5"))
	start := time.Now()
	if err := p.Start(); err != nil {
		t.FailNow()
	}
	if err := p.Wait(); err != nil {
		t.FailNow()
	}
	if time.Now().Sub(start).Seconds() < 4 {
		t.FailNow()
	}
}

func TestStdout(t *testing.T) {
	say := "konichiwa"
	p := NewProcess(exec.Command("echo", say))
	if err := p.Start(); err != nil {
		t.FailNow()
	}
	if err := p.Wait(); err != nil {
		t.FailNow()
	}
	if strings.TrimSpace(p.Stdout()) != say {
		t.FailNow()
	}
}
