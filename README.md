# go-runcmd

`go-runcmd` is a Go library and common interface for running local and remote commands providing the Runner interface which helps to abstract away running local and remote shell commands

> Borrowed from [kovetskiy/runcmd-](https://github.com/kovetskiy/runcmd-)

## Install

```#!console
go get github.com/aucloud/go-runcmd
```

## Usage

First, import the library:

```#!go
import "github.com/aucloud/go-runcmd"
```

Next, create a runner: this is a type, that holds:

- for local commands: empty struct
- for remote commands: connecttion to a remote host;
  so, you can create only one remote runner to remote host

Local Runner:

```#!go
uunner, err := runcmd.NewLocalRunner()
if err != nil {
  //handle error
}
```

Remote Runner:

```#!go
runner, err := runcmd.NewRemoteKeyAuthRunner(
  "user",
  "127.0.0.1:22",
  "/home/user/id_rsa",
)
if err != nil {
  //handle error
}
```

After that, create a command, and call `.Run()`:

```#!go
c, err := runner.Command("date")
if err != nil {
  //handle error
}
out, err := c.Run()
if err != nil {
  //handle error
}
```

Both local and remote runners implements the `Runner` interface,
so, you can work with them as Runner:

```go
func listSomeDir(r Runner) error {
  c, err := r.Command("ls -la")
  if err != nil {
    //handle error
  }
  out, err := c.Run()
  if err != nil {
    //handle error
  }
  for _, i := range out {
    fmt.Println(i)
  }
}
```

Another useful code snippet: pipe from local to remote command:

```#!go
lRunner, err := NewLocalRunner()
if err != nil {
  //handle error
}

rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
if err != nil {
  //handle error
}

cLocal, err := lRunner.Command("date")
if err != nil {
  //handle error
}
if err = cmdLocal.Start(); err != nil {
  //handle error
}
cRemote, err := rRunner.Command("tee /tmp/tmpfile")
if err != nil {
  //handle error
}
if err = cRemote.Start(); err != nil {
  //handle error
}
if _, err = io.Copy(cRemote.StdinPipe(),cLocal.StdoutPipe(),); err != nil {
  //handle error
}

// Correct handle end of copying:
cmdLocal.Wait()
cmdRemote.StdinPipe().Close()
cmdRemote.Wait()
```

## License

`go-runcmd` is licensed under the terms of the [AGPLv3](/LICENSE)