[![Build status](https://github.com/k0sproject/footloose/actions/workflows/go.yml)]
[![Go Report Card](https://goreportcard.com/badge/github.com/k0sproject/footloose)](https://goreportcard.com/report/github.com/k0sproject/footloose)
[![GoDoc](https://godoc.org/github.com/k0sproject/footloose?status.svg)](https://godoc.org/github.com/k0sproject/footloose)

# footloose

`footloose` creates containers that look like virtual machines. Those
containers run `systemd` as PID 1 and a ssh daemon that can be used to login
into the container. Such "machines" behave very much like a VM, it's even
possible to run [`dockerd` in them][readme-did] :)

`footloose` can be used for a variety of tasks, wherever you'd like virtual
machines but want fast boot times or need many of them. An easy way to think
about it is: [Vagrant](https://www.vagrantup.com/), but with containers.

`footloose` in action:

[![asciicast](https://asciinema.org/a/226185.svg)](https://asciinema.org/a/226185)

[readme-did]: ./examples/docker-in-docker/README.md

## Install

### From source

Build and install `footloose` from source. It requires having
`go >= 1.11` installed:

```console
go get github.com/k0sproject/footloose
```

[gh-release]: https://github.com/k0sproject/footloose/releases

## Usage

`footloose` reads a description of the *Cluster* of *Machines* to create from a
file, by default named `footloose.yaml`. An alternate name can be specified on
the command line with the `--config` option or through the `FOOTLOOSE_CONFIG`
environment variable.

The `config` command helps with creating the initial config file:

```console
# Create a footloose.yaml config file. Instruct we want to create 3 machines.
footloose config create --replicas 3
```

Start the cluster:

```console
$ footloose create
INFO[0000] Pulling image: quay.io/footloose/centos7 ...
INFO[0007] Creating machine: cluster-node0 ...
INFO[0008] Creating machine: cluster-node1 ...
INFO[0008] Creating machine: cluster-node2 ...
```

> It only takes a second to create those machines. The first time `create`
runs, it will pull the docker image used by the `footloose` containers so it
will take a tiny bit longer.

SSH into a machine with:

```console
$ footloose ssh root@node1
[root@1665288855f6 ~]# ps fx
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:00 /sbin/init
   23 ?        Ss     0:00 /usr/lib/systemd/systemd-journald
   58 ?        Ss     0:00 /usr/sbin/sshd -D
   59 ?        Ss     0:00  \_ sshd: root@pts/1
   63 pts/1    Ss     0:00      \_ -bash
   82 pts/1    R+     0:00          \_ ps fx
   62 ?        Ss     0:00 /usr/lib/systemd/systemd-logind
```

## Choosing the OS image to run

`footloose` will default to running a centos 7 container image. The `--image`
argument of `config create` can be used to configure the OS image. Valid OS
images are:

- `quay.io/footloose/centos7`
- `quay.io/footloose/fedora29`
- `quay.io/footloose/ubuntu16.04`
- `quay.io/footloose/ubuntu18.04`
- `quay.io/footloose/ubuntu20.04`
- `quay.io/footloose/amazonlinux2`
- `quay.io/footloose/debian10`
- `quay.io/footloose/clearlinux`

For example:

```console
footloose config create --replicas 3 --image quay.io/footloose/fedora29
```

Ubuntu images need the `--privileged` flag:

```console
footloose config create --replicas 1 --image quay.io/footloose/ubuntu16.04 --privileged
```

## `footloose.yaml`

`footloose config create` creates a `footloose.yaml` configuration file that is then
used by subsequent commands such as `create`, `delete` or `ssh`. If desired,
the configuration file can be named differently and supplied with the
`-c, --config` option.

```console
$ footloose config create --replicas 3
$ cat footloose.yaml
cluster:
  name: cluster
  privateKey: cluster-key
machines:
- count: 3
  backend: docker
  spec:
    image: quay.io/footloose/centos7
    name: node%d
    portMappings:
    - containerPort: 22
```


This configuration can naturally be edited by hand. The full list of
available parameters are in [the reference documentation][pkg-config].

[pkg-config]: https://godoc.org/github.com/k0sproject/footloose/pkg/config

## Examples

Interesting things can be done with `footloose`!

- [Customize the OS image](./examples/fedora29-htop/README.md)
- [Run Apache](./examples/apache/README.md)
- [Specify which ports on the hosts should be bound to services](examples/simple-hostPort/README.md)
- [Use Ansible to provision machines](./examples/ansible/README.md)
- [Run Docker inside `footloose` machines!](./examples/docker-in-docker/README.md)
- [Isolation and DNS resolution with custom docker networks](./examples/user-defined-network/README.md)
- [OpenShift with footloose](https://github.com/carlosedp/openshift-on-footloose)

## Under the hood

Under the hood, *Container Machines* are just containers. They can be
inspected with `docker`:

```console
$ docker ps
CONTAINER ID    IMAGE                        COMMAND         NAMES
04c27967f76e    quay.io/footloose/centos7    "/sbin/init"    cluster-node2
1665288855f6    quay.io/footloose/centos7    "/sbin/init"    cluster-node1
5134f80b733e    quay.io/footloose/centos7    "/sbin/init"    cluster-node0
```

The container names are derived from `cluster.name` and
`cluster.machines[].name`.

They run `systemd` as PID 1, it's even possible to inspect the boot messages:

```console
$ docker logs cluster-node1
systemd 219 running in system mode.
Detected virtualization docker.
Detected architecture x86-64.

Welcome to CentOS Linux 7 (Core)!

Set hostname to <1665288855f6>.
Initializing machine ID from random generator.
Failed to install release agent, ignoring: File exists
[  OK  ] Created slice Root Slice.
[  OK  ] Created slice System Slice.
[  OK  ] Reached target Slices.
[  OK  ] Listening on Journal Socket.
[  OK  ] Reached target Local File Systems.
         Starting Create Volatile Files and Directories...
[  OK  ] Listening on Delayed Shutdown Socket.
[  OK  ] Reached target Swap.
[  OK  ] Reached target Paths.
         Starting Journal Service...
[  OK  ] Started Create Volatile Files and Directories.
[  OK  ] Started Journal Service.
[  OK  ] Reached target System Initialization.
[  OK  ] Started Daily Cleanup of Temporary Directories.
[  OK  ] Reached target Timers.
[  OK  ] Listening on D-Bus System Message Bus Socket.
[  OK  ] Reached target Sockets.
[  OK  ] Reached target Basic System.
         Starting OpenSSH Server Key Generation...
         Starting Cleanup of Temporary Directories...
[  OK  ] Started Cleanup of Temporary Directories.
[  OK  ] Started OpenSSH Server Key Generation.
         Starting OpenSSH server daemon...
[  OK  ] Started OpenSSH server daemon.
[  OK  ] Reached target Multi-User System.
```

## FAQ

### Is `footloose` just like LXD?
In principle yes, but it will also work with Docker container images and
on MacOS as well.

## Help

We are a very friendly community and love questions, help and feedback.

If you have any questions, feedback, or problems with `footloose`:

- Check out the [examples](examples).
- [File an issue](https://github.com/k0sproject/footloose/issues/new).

footloose follows the [CNCF Code of
Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported by contacting a footloose project maintainer, or Natanael Copa
<ncopa@mirantis.com>

Your feedback is always welcome!
