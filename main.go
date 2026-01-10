package main

import "github.com/alt-dima/iacconsole-cli/cmd"

var (
	version string = "undefined"
)

func main() {
	cmd.SetVersionInfo(version)
	cmd.Execute()
}
