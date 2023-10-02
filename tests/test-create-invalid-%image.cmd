# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
# Test that bootloose create fails when the config is invalid (missing node number %d)

%defer bootloose delete --config test-create-invalid.static
%error bootloose create --config test-create-invalid.static
