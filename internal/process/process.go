package process

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
)

const (
	Pending = iota
	Running
	Error
	Exited
)

type State uint8

func (s State) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Running:
		return "Running"
	case Error:
		return "Error"
	case Exited:
		return "Exited"
	default:
		return ""
	}
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Process struct {
	cmd  *exec.Cmd
	exit chan error
	wg   sync.WaitGroup

	stdout string
	stderr string
}

func NewProcess(cmd *exec.Cmd) *Process {
	p := &Process{
		cmd:  cmd,
		exit: make(chan error, 1),
	}
	return p
}

func (p *Process) Start() error {
	// setup pipe before cmd start
	stdout, stderr, err := p.setupPipe()
	if err != nil {
		return err
	}
	err = p.cmd.Start()
	if err != nil {
		close(p.exit)
		return err
	}
	go func() {
		p.stdout = <-stdout
		p.stderr = <-stderr
		err := p.cmd.Wait()
		p.exit <- err
		close(p.exit)
	}()
	return nil
}

func (p *Process) Wait() error {
	return <-p.exit
}

func (p *Process) Cmd() string {
	return p.cmd.String()
}

func (p *Process) Stdout() string {
	return p.stdout
}

func (p *Process) Stderr() string {
	return p.stderr
}

func (p *Process) setupPipe() (chan string, chan string, error) {
	op, err := p.cmd.StdoutPipe()
	if err != nil {
		logrus.Warn("fail to read stdout", "error", err, "cmd", p.cmd.String())
		return nil, nil, err
	}
	ep, err := p.cmd.StderrPipe()
	if err != nil {
		logrus.Warn("fail to read stderr", "error", err, "cmd", p.cmd.String())
		return nil, nil, err
	}
	var (
		stdout = make(chan string, 1)
		stderr = make(chan string, 1)
	)
	go readPipe(op, stdout)
	go readPipe(ep, stderr)
	return stdout, stderr, nil
}

func readPipe(rc io.ReadCloser, ch chan string) {
	defer close(ch)
	bytes, err := ioutil.ReadAll(rc)
	if err != nil {
		logrus.Warn("fail to read pipe", "error", err)
		ch <- ""
		return
	}
	ch <- string(bytes)
}
