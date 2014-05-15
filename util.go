// utility functions
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
)

func DirName() string {
	p, _ := os.Getwd()
	return path.Base(p)
}

func PkgName(name string) string {
	binName := name
	if binName == "" {
		binName = DirName()
	}
	return fmt.Sprintf(".dist/%s-%s-%s-%s", binName, current_version, runtime.GOOS, runtime.GOARCH)
}

func IsGitRepo() bool {
	err := exec.Command("git", "rev-parse", "--git-dir").Run()
	return err == nil
}

func CaptureLogs(name string, c *exec.Cmd) bool {
	fmt.Fprintf(logFile, "\n---\n%s\n---\n", name)
	fmt.Fprintf(errLogFile, "\n---\n%s\n---\n", name)
	c.Stdout = logFile
	c.Stderr = errLogFile
	err := c.Start()
	if err != nil {
		return false
	}
	c.Wait()
	return true
}
