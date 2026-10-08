//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const installScript = `#!/bin/bash
set -eu
pid="$1"
src="$2"
dest="$3"
i=0
while kill -0 "$pid" 2>/dev/null; do
  i=$((i+1))
  if [ "$i" -gt 150 ]; then
    break
  fi
  sleep 0.2
done
sleep 0.4
rm -rf "$dest"
mv "$src" "$dest"
xattr -dr com.apple.quarantine "$dest" 2>/dev/null || true
open "$dest"
`

func stageAndRelaunch(stagedApp, bundle string) error {
	if _, err := os.Stat(filepath.Join(stagedApp, "Contents", "MacOS", "LifeOS")); err != nil {
		return fmt.Errorf("update bundle: %w", err)
	}
	script := filepath.Join(filepath.Dir(stagedApp), "install-update.sh")
	if err := os.WriteFile(script, []byte(installScript), 0o700); err != nil {
		return err
	}
	cmd := exec.Command(script, strconv.Itoa(os.Getppid()), stagedApp, bundle)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
