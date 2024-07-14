package version

import (
	"runtime"
	"runtime/debug"
)

const Name = "speedtest-exporter"

var version = "devel"

func Version() string {
	var commit string
	buildinfo, _ := debug.ReadBuildInfo()
	for _, item := range buildinfo.Settings {
		if item.Key == "vcs.revision" {
			commit = item.Value
			break
		}
	}
	if len(commit) > 7 {
		commit = commit[:7]
	} else if commit == "" {
		commit = "Unknown"
	}

	result := Name + ":\n"
	result += "    Version: " + version + "\n"
	result += "    Commit:  " + commit + "\n"
	result += "    Go:      " + runtime.Version() + "\n"

	return result
}
