# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM ubuntu18.04

RUN apt-get update && apt-get install -y apache2
COPY index.html /var/www/html

RUN systemctl enable apache2.service

EXPOSE 80
