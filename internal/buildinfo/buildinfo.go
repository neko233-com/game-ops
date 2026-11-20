// Package buildinfo carries version metadata injected at build time.
package buildinfo

import "runtime"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Set(v, c, d string) {
	if v != "" {
		version = v
	}
	if c != "" {
		commit = c
	}
	if d != "" {
		date = d
	}
}

func Version() string { return version }
func Commit() string  { return commit }
func Date() string    { return date }
func GoVersion() string { return runtime.Version() }
