package _exec

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

func NewSSHClient(server *Server) (client *ssh.Client, err error) {
	config, err := NewSSHClientConfig(server)
	if err != nil {
		return
	}
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), config)

	if err != nil {
		return
	}
	return
}

func NewSSHClientConfig(server *Server) (config *ssh.ClientConfig, err error) {

	var auth []ssh.AuthMethod

	if server.Password != "" {
		auth = append(auth, ssh.Password(server.Password))
	}

	if server.IdentityFile != "" {
		var ident ssh.AuthMethod
		ident, err = NewSSHPublicKey(server.IdentityFile)
		if err != nil {
			return
		}
		auth = append(auth, ident)
	}

	config = &ssh.ClientConfig{
		User:            server.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         60 * time.Second,
	}

	return
}

func NewProxySSHClient(proxyAddress string, server *Server) (client *ssh.Client, err error) {

	config, err := NewSSHClientConfig(server)
	if err != nil {
		return
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	serverAddress := fmt.Sprintf("%s:%d", server.Host, server.Port)
	conn, err := dialer.Dial("tcp", serverAddress)
	if err != nil {
		return
	}

	c, channel, reqs, err := ssh.NewClientConn(conn, serverAddress, config)
	if err != nil {
		return
	}

	return ssh.NewClient(c, channel, reqs), nil
}

func NewSSHPublicKey(path string) (auth ssh.AuthMethod, err error) {

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
