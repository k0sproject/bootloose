<!--
SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
SPDX-FileCopyrightText: 2023 bootloose authors
SPDX-License-Identifier: Apache-2.0
-->
# Run Apache with bootloose

Using the bootloose base like above, create a docker file which installs Apache and
exposes a port like 80 or 443:

```Dockerfile
FROM ubuntu18.04

RUN apt-get update && apt-get install -y apache2
COPY index.html /var/www/html

RUN systemctl enable apache2.service

EXPOSE 80
```

Build that image:

```console
$ docker built -t apache:test01 .
```

Create a bootloose configuration file.

```console
$ bootloose config create --image apache:test01
```

Now, create a machine!

```console
$ bootloose create
```

Once the machine is ready, you should be able to access apache on the exposed port.

```console
$ docker port cluster-node0 80
0.0.0.0:32824
$ curl 0.0.0.0:32824
<!DOCTYPE html>
<html>
    <title>bootloose</title>
    <body>
        Hello, from bootloose!
    </body>
</html>
```

In case of multiple machines the port will be different on each machine.

```console
$ docker port cluster-node1 80
0.0.0.0:32828

$ docker port cluster-node0 80
0.0.0.0:32826
```
