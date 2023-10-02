# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM quay.io/k0sproject/bootloose-fedora38

# Pre-seed the htop package
RUN dnf -y install htop && dnf clean all
