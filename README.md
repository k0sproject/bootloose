<!--
SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
SPDX-FileCopyrightText: 2023 bootloose authors
SPDX-License-Identifier: Apache-2.0
-->
[![Go](https://github.com/k0sproject/bootloose/actions/workflows/go.yaml/badge.svg)](https://github.com/k0sproject/bootloose/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/k0sproject/bootloose)](https://goreportcard.com/report/github.com/k0sproject/bootloose)
[![GoDoc](https://godoc.org/github.com/k0sproject/bootloose?status.svg)](https://godoc.org/github.com/k0sproject/bootloose)

# bootloose

`bootloose` creates containers that look like virtual machines. Those
containers run `systemd` as PID 1 and a ssh daemon that can be used to login
into the container. Such "machines" behave very much like a VM, it's even
possible to run [`dockerd` in them](./examples/docker-in-docker/).

`bootloose` can be used for a variety of tasks, wherever you'd like virtual
machines but want fast boot times or need many of them. An easy way to think
about it is: [Vagrant](https://www.vagrantup.com/), but with containers.

`bootloose` in action:

```console
$ bootloose config create --replicas 3
$ bootloose create
INFO[0000] Pulling image: quay.io/k0sproject/bootloose-ubuntu20.04 ...
INFO[0007] Creating machine: cluster-node0 ...
INFO[0008] Creating machine: cluster-node1 ...
INFO[0008] Creating machine: cluster-node2 ...
$ docker ps
CONTAINER ID    IMAGE                                     COMMAND         NAMES
04c27967f76e    quay.io/k0sproject/bootloose-ubuntu20.04  "/sbin/init"    cluster-node2
1665288855f6    quay.io/k0sproject/bootloose-ubuntu20.04  "/sbin/init"    cluster-node1
5134f80b733e    quay.io/k0sproject/bootloose-ubuntu20.04  "/sbin/init"    cluster-node0
$ bootloose ssh root@node1
[root@1665288855f6 ~]# █
```

## Attribution

This project is a continuation of [Footloose](https://github.com/weaveworks/footloose) by Weaveworks. 
We are grateful for their work and contributions from the community.

## Install

### Homebrew

Install using [Homebrew](https://brew.sh/) package manager:

```console
brew install k0sproject/tap/bootloose
```

### From source

Build and install `bootloose` from source. It requires having
`go >= 1.21` installed:

```console
go install github.com/k0sproject/bootloose@latest
```

[gh-release]: https://github.com/k0sproject/bootloose/releases

## Usage

`bootloose` reads a description of the *Cluster* of *Machines* to create from a
file, by default named `bootloose.yaml`. An alternate name can be specified on
the command line with the `--config` option or through the `BOOTLOOSE_CONFIG`
environment variable.

The `config` command helps with creating the initial config file:

```console
# Create a bootloose.yaml config file. Instruct we want to create 3 machines.
bootloose config create --replicas 3
```

Start the cluster:

```console
$ bootloose create
INFO[0000] Pulling image: quay.io/k0sproject/bootloose-debian12 ...
INFO[0007] Creating machine: cluster-node0 ...
INFO[0008] Creating machine: cluster-node1 ...
INFO[0008] Creating machine: cluster-node2 ...
```

> It only takes a second to create those machines. The first time `create`
runs, it will pull the docker image used by the `bootloose` containers so it
will take a tiny bit longer.

SSH into a machine with:

```console
$ bootloose ssh root@node1
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

`bootloose` will default to running an Ubuntu LTS container image. The `--image`
argument of `config create` can be used to configure the OS image. OS
images provided by this repository are:

- `quay.io/k0sproject/bootloose-alpine3.18`
- `quay.io/k0sproject/bootloose-amazonlinux2023`
- `quay.io/k0sproject/bootloose-amazonlinux2`
- `quay.io/k0sproject/bootloose-clearlinux`
- `quay.io/k0sproject/bootloose-debian10`
- `quay.io/k0sproject/bootloose-debian12`
- `quay.io/k0sproject/bootloose-fedora38`
- `quay.io/k0sproject/bootloose-rockylinux9`
- `quay.io/k0sproject/bootloose-ubuntu18.04`
- `quay.io/k0sproject/bootloose-ubuntu20.04`
- `quay.io/k0sproject/bootloose-ubuntu22.04`

The tag `:latest` is updated when any of the images are changed in the repository.
When bootloose CLI binary releases are published, images at that point are tagged
with a version that you can pin a config to, such as
`quay.io/k0sproject/bootloose-ubuntu20.04:v0.7.0`.

For example:

```console
bootloose config create --replicas 3 --image quay.io/k0sproject/bootloose-debian12
```

```console
bootloose config create --replicas 3 --image quay.io/k0sproject/bootloose-debian12:v0.7.0
```

Some images may need the `--privileged` flag.

## `bootloose.yaml`

`bootloose config create` creates a `bootloose.yaml` configuration file that is then
used by subsequent commands such as `create`, `delete` or `ssh`. If desired,
the configuration file can be named differently and supplied with the
`-c, --config` option.

```console
$ bootloose config create --replicas 3
$ cat bootloose.yaml
cluster:
  name: cluster
  privateKey: cluster-key
machines:
- count: 3
  spec:
    image: quay.io/k0sproject/bootloose-debian12
    name: node%d
    portMappings:
    - containerPort: 22
```


This configuration can naturally be edited by hand. The full list of
available parameters are in [the reference documentation][pkg-config].

[pkg-config]: https://godoc.org/github.com/k0sproject/bootloose/pkg/config

## Examples

Interesting things can be done with `bootloose`!

- [Customize the OS image](./examples/fedora38-htop/README.md)
- [Run Apache](./examples/apache/README.md)
- [Specify which ports on the hosts should be bound to services](examples/simple-hostPort/README.md)
- [Use Ansible to provision machines](./examples/ansible/README.md)
- [Run Docker inside `bootloose` machines!](./examples/docker-in-docker/README.md)
- [Isolation and DNS resolution with custom docker networks](./examples/user-defined-network/README.md)
- [OpenShift with bootloose](https://github.com/carlosedp/openshift-on-bootloose)

## Under the hood

Under the hood, *Container Machines* are just containers. They can be
inspected with `docker`:

```console
$ docker ps
CONTAINER ID    IMAGE                                  COMMAND         NAMES
04c27967f76e    quay.io/k0sproject/bootloose-debian12  "/sbin/init"    cluster-node2
1665288855f6    quay.io/k0sproject/bootloose-debian12  "/sbin/init"    cluster-node1
5134f80b733e    quay.io/k0sproject/bootloose-debian12  "/sbin/init"    cluster-node0
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

### Is `bootloose` just like LXD?
In principle yes, but it will also work with Docker container images and
on MacOS as well.

## Help

We are a very friendly community and love questions, help and feedback.

If you have any questions, feedback, or problems with `bootloose`:

- Check out the [examples](examples).
- [File an issue](https://github.com/k0sproject/bootloose/issues/new).
- [Contact us](https://k0sproject.io/contact-us.html) via the form on the k0sproject website.

bootloose follows the [CNCF Code of
Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported by contacting a bootloose project maintainer.

