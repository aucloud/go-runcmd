# runcmd

runcmd is a common interface for running local and remote commands and provides
the `Runner` interface which helps to abstract away running local and remote
shell commands.

> Borrowed from https://github.com/kovetskiy/runcmd-

## Install

```#!console
go get gitlab.mgt.aom.australiacloud.com.au/aom/golib/runcmd
```

## Usage

First, create runner: this is a type, that holds:

- for local commands: empty struct
- for remote commands: connect to remote host;
  so, you can create only one remote runner to remote host

```#!go
lRunner, err := runcmd.NewLocalRunner()
if err != nil {
  //handle error
}

rRunner, err := runcmd.NewRemoteKeyAuthRunner(
  "user",
  "127.0.0.1:22",
  "/home/user/id_rsa",
)
if err != nil {
  //handle error
}
```

After that, create command, and run methods:

```#!go
c, err := rRunner.Command("date")
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

// List some dir on local host:
if err := listSomeDir(lRunner); err != nil {
  //handle error
}

// List some dir on remote host:
if err := listSomeDir(rRunner); err != nil {
  //handle error
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
