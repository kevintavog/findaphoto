package main

import (
	"fmt"
)

var majorVersion = 1
var minorVersion = 0
var buildVersion = 1

func versionString() string {
	return fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, buildVersion)
}
