<!--
SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
SPDX-FileCopyrightText: 2023 bootloose authors
SPDX-License-Identifier: Apache-2.0
-->
# Ansible provisioned machine

create a new environment configuration:

```console
$ bootloose config create --replicas 1
```

deploy container images:

```console
$ bootloose create
INFO[0000] Pulling image: quay.io/bootloose/ubuntu20.04 ...
INFO[0007] Creating machine: cluster-node0 ...
```


test the ansible setup:

```console
$ ansible -m ping all
cluster-node0 | SUCCESS => {
    "changed": false, 
    "ping": "pong"
}
```

run the provisioning playbook:

```console
$ ansible-playbook  example1.yml

PLAY [Install nginx] ****************************

TASK [Gathering Facts] **************************
ok: [cluster-node0]

TASK [Add epel-release repo] ********************
changed: [cluster-node0]

TASK [Install nginx] ****************************
changed: [cluster-node0]

TASK [Insert Index Page] ************************
changed: [cluster-node0]

TASK [Start NGiNX] ******************************
changed: [cluster-node0]

PLAY RECAP **************************************
cluster-node0              : ok=5    changed=0    unreachable=0    failed=0
```
