package version

import (
	"github.com/carlmjohnson/versioninfo"
)

var (
	// Version of the product, is set during the build
	Version = versioninfo.Version
	// GitCommit is set during the build
	GitCommit = versioninfo.Revision
	// Environment of the product, is set during the build
	Environment = "development"
)
