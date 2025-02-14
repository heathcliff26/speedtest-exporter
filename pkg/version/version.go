package version

import (
	"runtime"
	"runtime/debug"
)

const Name = "speedtest-exporter"

// Return a formated string containing the version, git commit and go version the app was compiled with.
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
	result += "    Version: " + buildinfo.Main.Version + "\n"
	result += "    Commit:  " + commit + "\n"
	result += "    Go:      " + runtime.Version() + "\n"

	return result
}
