package runcmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"testing"
)

var (
	cmdValid      = "ls -la"
	cmdValidArgs  = `ls "-la"`
	cmdInvalid    = "blah-blah"
	cmdInvalidKey = "uname -blah"
	cmdPipeOut    = "date"
	cmdPipeIn     = "/usr/bin/tee /tmp/blah"
)

/* FIXME: Mock an SSH server
func TestKeyAuth(t *testing.T) {
	rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	if err := testRun(rRunner); err != nil {
		t.Error(err)
	}
}
*/

/* FIXME: Mock an SSH server with password auth
func TestPassAuth(t *testing.T) {
	defer func() {
		if er := recover(); er != nil {
			os.Exit(1)
		}
	}()
	rRunner, err := NewRemotePassAuthRunner(user, host, pass)
	if err != nil {
		t.Error(err)
	}
	if err := testRun(rRunner); err != nil {
		t.Error(err)
	}
}
*/

func TestLocalRun(t *testing.T) {
	lRunner, err := NewLocalRunner()
	if err != nil {
		t.Error(err)
	}
	if err := testRun(lRunner); err != nil {
		t.Error(err)
	}
}

/* FIXME: Mock anSSH server
func TestRemoteRun(t *testing.T) {
	rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	if err := testRun(rRunner); err != nil {
		t.Error(err)
	}
}
*/

func TestLocalStartWait(t *testing.T) {
	lRunner, err := NewLocalRunner()
	if err != nil {
		t.Error(err)
	}
	if err := testStartWait(lRunner); err != nil {
		t.Error(err)
	}
}

/* FIXME: Mock an SSH server
func TestRemoteStartWait(t *testing.T) {
	rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	if err := testStartWait(rRunner); err != nil {
		t.Error(err)
	}
}

func TestPipeLocal2Remote(t *testing.T) {
	if err := testPipe(true); err != nil {
		t.Error(err)
	}
}

func TestPipeRemote2Local(t *testing.T) {
	if err := testPipe(false); err != nil {
		t.Error(err)
	}
}
*/

func testRun(runner Runner) error {
	// Valid command with valid keys:
	cmd, err := runner.Command(cmdValid)
	if err != nil {
		return err
	}
	out, err := cmd.Run()
	if err != nil {
		return err
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with valid keys:
	// (Arguments are quoted)
	cmd, err = runner.Command(cmdValidArgs)
	if err != nil {
		return err
	}
	out, err = cmd.Run()
	if err != nil {
		return err
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with invalid keys:
	cmd, err = runner.Command(cmdInvalidKey)
	if err != nil {
		return err
	}
	if _, err = cmd.Run(); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd, err = runner.Command(cmdInvalid)
	if err != nil {
		return err
	}
	if _, err = cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

func testStartWait(runner Runner) error {
	// Valid command with valid keys:
	cmd, err := runner.Command(cmdValid)
	if err != nil {
		return err
	}
	b, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	bOut, err := ioutil.ReadAll(b)
	for _, s := range strings.Split(strings.Trim(string(bOut), "\n"), "\n") {
		fmt.Println(s)
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	// Valid command with invalid keys:
	cmd, err = runner.Command(cmdInvalidKey)
	if err != nil {
		return err
	}
	b, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	bOut, err = ioutil.ReadAll(b)
	for _, s := range strings.Split(strings.Trim(string(bOut), "\n"), "\n") {
		fmt.Println(s)
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd, err = runner.Command(cmdInvalid)
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

/* FIXME: Mock an SSH server
func testPipe(localToRemote bool) error {
	lRunner, err := NewLocalRunner()
	if err != nil {
		return err
	}
	rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
	if err != nil {
		return err
	}

	if localToRemote {
		cmdLocal, err := lRunner.Command(cmdPipeOut)
		if err != nil {
			return err
		}
		localStdout, err := cmdLocal.StdoutPipe()
		if err != nil {
			return err
		}
		if err = cmdLocal.Start(); err != nil {
			return err
		}
		cmdRemote, err := rRunner.Command(cmdPipeIn)
		if err != nil {
			return err
		}
		remoteStdin, err := cmdRemote.StdinPipe()
		if err != nil {
			return err
		}
		if err = cmdRemote.Start(); err != nil {
			return err
		}
		if _, err = io.Copy(remoteStdin, localStdout); err != nil {
			return err
		}
		err = remoteStdin.Close()
		if err != nil {
			return err
		}
		return cmdLocal.Wait()
	}

	cmdLocal, err := lRunner.Command(cmdPipeIn)
	if err != nil {
		return err
	}
	localStdin, err := cmdLocal.StdinPipe()
	if err != nil {
		return err
	}
	if err = cmdLocal.Start(); err != nil {
		return err
	}
	cmdRemote, err := rRunner.Command(cmdPipeOut)
	if err != nil {
		return err
	}
	remoteStdout, err := cmdRemote.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmdRemote.Start(); err != nil {
		return err
	}
	if _, err = io.Copy(localStdin, remoteStdout); err != nil {
		return err
	}
	err = localStdin.Close()
	if err != nil {
		return err
	}
	return cmdRemote.Wait()
}
*/
