//go:build !windows

package main

import "os"

var iconData []byte

func startTray(url, tip string, ready chan<- bool) { ready <- false }

func exitApp() { logf("arrêt"); os.Exit(0) }
