<!--
SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
SPDX-FileCopyrightText: 2023 bootloose authors
SPDX-License-Identifier: Apache-2.0
-->
# End-to-end tests

This directory holds `bootloose` end-to-end tests. All commands given in this
README are assuming being run from the root of the repository.

The prerequisites to run to the tests are:

- `docker` installed on the machine with no container running. This
limitation can be lifted once we can select `bootloose` containers better
([#17][issue-17]).
- `bootloose` in the path.

[issue-17]: https://github.com/weaveworks/footloose/issues/17

## Running the tests

To run all tests:

```console
go test -v ./tests
```

To exclude long running tests (useful to smoke test a change before a longer
run in CI):

```console
go test -short -v ./tests
```

To run a specific test:

```console
go test -v -run TestEndToEnd/test-create-delete-centos7
```

Remember that the `-run` argument is a regex so it's possible to select a
subset of the tests with this:

```console
go test -v -run TestEndToEnd/test-create-delete
```

This will match `test-create-delete-centos7`, `test-create-delete-fedora38`,
...

To run tests for a specific image:

```console
go test -v -args -image=image_name ./tests
```

## Writing tests

`bootloose` has a small framework to write end to end tests. The main idea is
to write a `.cmd` file with a list of commands to run and compare the output
(stdout+stderr) of those commands to a golden, expected, output.

`.cmd` files look like (`test-ssh-remote-command-%image.cmd`):

```shell
# Test bootloose ssh can execute a remote command
bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image
bootloose create --config %testName.bootloose
%out bootloose --config %testName.bootloose ssh root@node0 hostname
bootloose delete --config %testName.bootloose
```

And the corresponding golden output file (`test-ssh-remote-command-%image.golden.output`):

```shell
node0
```

The **--override** flag should be used with the **config create** command because otherwise 
the first run of a test will leave a config file behind and additional runs will fail
to avoid overwriting the original config file. The only exception to this rule is in tests
that are intended to validate the override mechanism itself.

Some variables and directives are supplied by the test framework:

- **%testName**: The name of the test. This is really the name of the `.cmd`
file without the extension.

- **%out**: Capture the output of the following command to be compared to the
golden output. In the example above the result of the remote `hostname`
command will be compared to `node0`.

