//go:build windows

package main

import "fmt"

func stageAndRelaunch(stagedApp, bundle string) error {
	return fmt.Errorf("mac app update is not applied on windows (%s -> %s)", stagedApp, bundle)
}
