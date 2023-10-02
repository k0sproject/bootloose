# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image ubuntu18.04
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose
bootloose create --config %testName.bootloose
bootloose delete --config %testName.bootloose
%out bootloose show --config %testName.bootloose
%out bootloose show -o json --config %testName.bootloose
