/*
	go-runcmd is a Go library and common interface for running local
	and remote commands providing the Runner interface which helps
	to abstract away running local and remote shell commands

    Copyright (C) 2021 Sovereign Cloud Australia Pty Ltd

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package runcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	sshutil "github.com/aucloud/go-sshutil"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type RemoteCmd struct {
	cmdline string
	session *ssh.Session
}

type Remote struct {
	client *sshutil.Client
}

func ResolveHostname(hostport string) (net.Addr, error) {
	var (
		host string
		port string
	)

	if strings.Contains(hostport, ":") {
		tokens := strings.SplitN(hostport, ":", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error parsing hostport %s, expected 2 tokens got %d", hostport, len(tokens))
		}
		host = tokens[0]
		port = tokens[1]
		if _, err := strconv.Atoi(port); err != nil {
			return nil, fmt.Errorf("error parsing hostport %s, expected <host>:<port> and <port> to be an int: %w", hostport, err)
		}
	} else {
		host = hostport
		port = "22"
	}

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func NewRemoteKeyAuthRunner(ctx context.Context, user, host, key string) (*Remote, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading private ssh key %s: %w", key, err)
	}
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("error reading private ssh key %s: %w", key, err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private ssh key %s: %w", key, err)
	}
	config := &ssh.ClientConfig{
		User: user,
		// FIXME: This is insecure. We should verify RSA fingerprints of hosts...
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	addr, err := ResolveHostname(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", host, err)
	}
	client, err := sshutil.NewClient(
		ctx,
		sshutil.ConstantAddrResolver{addr},
		config,
		sshutil.DefaultConnectBackoff(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish an SSH connection to %s: %w", host, err)
	}
	return &Remote{client}, nil
}

func NewRemoteAgentAuthRunner(ctx context.Context, user, host, agentSocket string) (*Remote, error) {

	if _, err := os.Stat(agentSocket); os.IsNotExist(err) {
		return nil, fmt.Errorf("agent socket %s does not exist: %w", agentSocket, err)
	}
	agentConn, err := net.Dial("unix", agentSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH agent socket %s: %v", agentSocket, err)
	}
	agentClient := agent.NewClient(agentConn)
	config := &ssh.ClientConfig{
		User: user,
		// FIXME: This is insecure. We should verify RSA fingerprints of hosts...
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)},
	}
	addr, err := ResolveHostname(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", host, err)
	}
	client, err := sshutil.NewClient(
		ctx,
		sshutil.ConstantAddrResolver{addr},
		config,
		sshutil.DefaultConnectBackoff(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish an SSH connection to %s: %w", host, err)
	}
	return &Remote{client}, nil
}

func NewRemoteKeyAuthRunnerViaJumphost(ctx context.Context, user, host, jumphost, key string) (*Remote, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading private ssh key %s: %w", key, err)
	}
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("error reading private ssh key %s: %w", key, err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private ssh key %s: %w", key, err)
	}
	config := &ssh.ClientConfig{
		User: user,
		// FIXME: This is insecure. We should verify RSA fingerprints of hosts...
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	// First connect to the Jumphost
	jumphostAddr, err := ResolveHostname(jumphost)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", jumphost, err)
	}
	bastionClient, err := sshutil.NewClient(
		ctx,
		sshutil.ConstantAddrResolver{jumphostAddr},
		config,
		sshutil.DefaultConnectBackoff(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish an SSH connection to %s: %w", host, err)
	}

	// Next connect to the target host
	addr, err := ResolveHostname(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", jumphost, err)
	}
	client, err := bastionClient.ConnectTo(
		ctx,
		sshutil.ConstantAddrResolver{addr},
		config,
		sshutil.DefaultConnectBackoff(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish an SSH connection to %s: %w", host, err)
	}

	return &Remote{client}, nil
}

func NewRemotePassAuthRunner(ctx context.Context, user, host, password string) (*Remote, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	addr, err := ResolveHostname(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %w", host, err)
	}
	client, err := sshutil.NewClient(
		ctx,
		sshutil.ConstantAddrResolver{addr},
		config,
		sshutil.DefaultConnectBackoff(),
	)
	return &Remote{client}, nil
}

func (runner *Remote) Command(cmdline string) (CmdWorker, error) {
	if cmdline == "" {
		return nil, errors.New("command cannot be empty")
	}

	session, err := runner.client.Client().NewSession()
	if err != nil {
		return nil, err
	}

	return &RemoteCmd{
		cmdline: cmdline,
		session: session,
	}, nil
}

func (runner *Remote) CloseConnection() error {
	runner.client.Close()
	return nil
}

func (cmd *RemoteCmd) Run() ([]string, error) {
	defer cmd.session.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return run(cmd)
}

func (cmd *RemoteCmd) Start() error {
	return cmd.session.Start(cmd.cmdline)
}

func (cmd *RemoteCmd) Wait() error {
	defer cmd.session.Close()

	return cmd.session.Wait()
}

func (cmd *RemoteCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.session.StdinPipe()
}

func (cmd *RemoteCmd) StdoutPipe() (io.Reader, error) {
	return cmd.session.StdoutPipe()
}

func (cmd *RemoteCmd) StderrPipe() (io.Reader, error) {
	return cmd.session.StderrPipe()
}

func (cmd *RemoteCmd) SetStdout(buffer io.Writer) {
	cmd.session.Stdout = buffer
}

func (cmd *RemoteCmd) SetStderr(buffer io.Writer) {
	cmd.session.Stderr = buffer
}

func (cmd *RemoteCmd) GetCommandLine() string {
	return cmd.cmdline
}
