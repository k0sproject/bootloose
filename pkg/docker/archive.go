/*
Copyright 2018 The Kubernetes Authors.
Copyright 2019 Weaveworks Ltd.
Copyright 2023 bootloose authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package docker contains helpers for working with docker
// This package has no stability guarantees whatsoever!
package docker

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// GetArchiveTags obtains a list of "repo:tag" docker image tags from a
// given docker image archive (tarball) path
// compatible with all known specs:
// https://github.com/moby/moby/blob/master/image/spec/v1.0.md
// https://github.com/moby/moby/blob/master/image/spec/v1.1.md
// https://github.com/moby/moby/blob/master/image/spec/v1.2.md
func GetArchiveTags(path string) ([]string, error) {
	// open the archive and find the repositories entry
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tr := tar.NewReader(f)
	var hdr *tar.Header
	for {
		hdr, err = tr.Next()
		if err == io.EOF {
			return nil, errors.New("could not find image metadata")
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == "repositories" {
			break
		}
	}
	// read and parse the tags
	b, err := io.ReadAll(tr)
	if err != nil {
		return nil, err
	}
	var repoTags map[string]map[string]string
	if err := json.Unmarshal(b, &repoTags); err != nil {
		return nil, err
	}
	res := []string{}
	for repo, tags := range repoTags {
		for tag := range tags {
			res = append(res, fmt.Sprintf("%s:%s", repo, tag))
		}
	}
	return res, nil
}
