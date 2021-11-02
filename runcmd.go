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
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ExecError struct {
	ExecutionError error
	CommandLine    string
	Output         []string
}

type Runner interface {
	Command(cmd string) (CmdWorker, error)
}

type CmdWorker interface {
	Run() ([]string, error)
	Start() error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	SetStdout(buffer io.Writer)
	SetStderr(buffer io.Writer)
	GetCommandLine() string
}

func newExecError(
	execErr error, cmdline string, output []string,
) ExecError {
	return ExecError{execErr, cmdline, output}
}

func (err ExecError) Error() string {
	errString := fmt.Sprintf(
		"`%s` failed: %s", err.CommandLine, err.ExecutionError,
	)

	output := strings.Join(err.Output, "\n")
	if strings.TrimSpace(output) != "" {
		errString = errString + ", output: \n" + output
	}

	return errString
}

func run(worker CmdWorker) ([]string, error) {
	var buffer bytes.Buffer

	worker.SetStdout(&buffer)
	worker.SetStderr(&buffer)

	err := worker.Wait()
	output := strings.Split(string(buffer.Bytes()), "\n")

	if err != nil {
		return nil, newExecError(err, worker.GetCommandLine(), output)
	}

	return output, nil
}
