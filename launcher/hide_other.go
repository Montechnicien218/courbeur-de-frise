//go:build !windows

package main

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

func hideWindow(c *exec.Cmd) {}

func shellOpen(target string) { exec.Command("xdg-open", target).Start() }

func openFolder(d string) { shellOpen(d) }

func alert(msg string) { println(msg) }

func uptime() time.Duration { return 0 }

func isFrench() bool {
	l := strings.ToLower(os.Getenv("LANG") + os.Getenv("LC_ALL") + os.Getenv("LANGUAGE"))
	return strings.HasPrefix(l, "fr")
}
