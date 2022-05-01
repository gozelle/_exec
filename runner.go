package _exec

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func NewRunner() *Runner {
	return &Runner{
		shell: "bash",
	}
}

type Runner struct {
	commands []string
	shell    string
	dir      string
	environ  []string
}

// AddCommand 设置执行命令
func (p *Runner) AddCommand(commands ...string) *Runner {
	p.commands = commands
	return p
}

// SetEnviron 设置环境变量
func (p *Runner) SetEnviron(environ []string) *Runner {
	p.environ = environ
	return p
}

// SetDir 设置运行目录
func (p *Runner) SetDir(path string) *Runner {
	p.shell = path
	return p
}

// SetShell 设置执行 Sell
func (p *Runner) SetShell(path string) *Runner {
	p.shell = path
	return p
}

// CombinedOutput 输出最终结果
func (p *Runner) CombinedOutput() (result string, err error) {
	var res []byte
	for _, v := range p.commands {
		c := exec.Command(p.shell, "-c", v)
		c.Dir = p.dir
		res, err = c.CombinedOutput()
		if err != nil {
			return
		}
		result += strings.TrimSpace(string(res))
	}
	return
}

// PipeOutput 即时输出结果
func (p *Runner) PipeOutput() {
	for _, v := range p.commands {
		p.pipeExec(v)
	}
}

func (p *Runner) wrapCmd(cmd *exec.Cmd) {
	cmd.Dir = p.dir
	if len(p.environ) > 0 {
		cmd.Env = p.environ
	} else {
		cmd.Env = os.Environ()
	}
}

func (p *Runner) pipeExec(command string) {
	c := exec.Command(p.shell, "-c", command)
	p.wrapCmd(c)
	c.Stdin = os.Stdin
	
	stderr, err := c.StderrPipe()
	if err != nil {
		return
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return
	}
	
	//c.Stderr = c.Stdout
	
	out := make(chan []byte)
	defer func() {
		close(out)
	}()
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	go func() {
		//enter := false
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			//if !enter {
			//	enter = true
			//	_, _ = os.Stderr.Write([]byte(fmt.Sprintf("%s \n", "错误")))
			//}
			_, _ = os.Stdout.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stderr.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	err = c.Run()
	wg.Wait()
	if err != nil {
		return
	}
}
