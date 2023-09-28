# Customize the OS image

It is possible to create docker images that specialize a [`bootloose` base
image](https://github.com/k0sproject/bootloose#choosing-the-os-image-to-run) to
suit your needs.

For instance, if we want the created machines to run `fedora38` with the
`htop` package already pre-installed:

```Dockerfile
FROM quay.io/k0sproject/bootloose-fedora39

# Pre-seed the htop package
RUN dnf -y install htop && dnf clean all

```

Build that image:

```console
docker build -t fedora39-htop .
```

Configure `bootloose.yaml` to use that image by either editing the file or running:

```console
bootloose config create --image fedora38-htop
````

`htop` will be available on the newly created machines!

```console
$ bootloose create
$ bootloose ssh root@node0
# htop
```
