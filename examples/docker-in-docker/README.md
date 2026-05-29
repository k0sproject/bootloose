<!--
SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
SPDX-FileCopyrightText: 2023 bootloose authors
SPDX-License-Identifier: Apache-2.0
-->
# Running `dockerd` in Container Machines

To run `dockerd` inside a docker container, two things are needed:

- Run the container as privileged (we could probably do better! expose
capabilities instead).
- Mount `/var/lib/containerd` as volume, here an anonymous volume. This is
because of [limitations][dind] of what you can do with the overlay system
docker is setup to use.

```yaml
cluster:
  name: cluster
  privateKey: cluster-key
machines:
- count: 1
  spec:
    image: quay.io/k0sproject/bootloose-ubuntu24.04
    name: node%d
    portMappings:
    - containerPort: 22
    privileged: true
    volumes:
    - type: volume
      destination: /var/lib/containerd
```

You can then install and run docker on the machine:

```console
$ bootloose create
$ bootloose ssh root@node0
# apt update && apt install -y docker.io
[...]
# systemctl start docker
# docker run busybox echo 'Hello, World!'
Hello, World!
```

[dind]: https://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/
