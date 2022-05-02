package _exec

import (
	"bufio"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Key          string
	Host         string
	Port         int
	User         string
	Password     string
	IdentityFile string
	sshClient    *ssh.Client
	sftpClient   *sftp.Client
}

func (p *Server) SSHClient() *ssh.Client {
	return p.sshClient
}

func (p *Server) SFTPClient() *sftp.Client {
	return p.sftpClient
}

func (p *Server) Close() {
	if p.sftpClient != nil {
		_ = p.sftpClient.Close()
	}
	if p.sshClient != nil {
		_ = p.sshClient.Close()
	}
}

func (p Server) getPublicKey(path string) (auth ssh.AuthMethod, err error) {

	home, err := HomeDir()
	if err != nil {
		return
	}

	if strings.HasPrefix(path, "~") {
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	key, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}
	auth = ssh.PublicKeys(signer)

	return

}

func (p *Server) InitSSH() (err error) {
	if p.sshClient == nil {
		p.sshClient, err = NewSSHClient(p)
		if err != nil {
			return
		}
	}
	return
}

func (p *Server) InitSFTP() (err error) {
	err = p.InitSSH()
	if err != nil {
		err = fmt.Errorf("connect server error: %s", err)
		return
	}
	if p.sftpClient == nil {
		p.sftpClient, err = sftp.NewClient(p.sshClient)
		if err != nil {
			err = fmt.Errorf("init sftp client error: %s", err)
			return
		}
	}
	return
}

func (p *Server) Ping() (err error) {
	err = p.InitSSH()
	if err != nil {
		return
	}
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return
	}
	defer func() {
		_ = session.Close()
	}()
	result, err := session.CombinedOutput(`echo "pong"`)
	if err != nil {
		err = fmt.Errorf("ping error: %s", err)
		return
	}
	ok := strings.TrimSpace(string(result))
	if ok != "pong" {
		err = fmt.Errorf("ping error: server not response 'pong'")
		return
	}
	return
}

func (p *Server) CombinedExec(command string) error {
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()
	out, err := session.CombinedOutput(command)
	if err != nil {
		err = fmt.Errorf(strings.TrimSpace(string(out)))
		return err
	}
	return nil
}

func (p *Server) PipeExec(command string) (err error) {
	err = p.InitSSH()
	if err != nil {
		return
	}
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return
	}
	defer func() {
		_ = session.Close()
	}()
	stderr, err := session.StderrPipe()
	if err != nil {
		err = fmt.Errorf("fetch stderr pipe error: %s", err)
		return
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("fetch stdout pipe error: %s", err)
		return
	}

	out := make(chan string, 1048576)
	defer func() {
		for len(out) > 0 {
			time.Sleep(500 * time.Millisecond)
			close(out)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
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

	err = session.Run(command)
	if err != nil {
		err = fmt.Errorf("session run command error: %s", err)
		return
	}
	wg.Wait()

	return
}
