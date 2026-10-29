//go:build windows

package main

import (
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

func hideWindow(c *exec.Cmd) { c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

// ouvre une adresse ou un dossier avec l'application par défaut de Windows
func shellOpen(target string) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	verb, _ := syscall.UTF16PtrFromString("open")
	t, _ := syscall.UTF16PtrFromString(target)
	r, _, _ := proc.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(t)), 0, 0, 1)
	logf("ShellExecute %s -> %d", target, r)
	if r <= 32 {
		logf("ouverture impossible de %s (code %d)", target, r)
		alert(L("Le Courbeur de frise est lancé, mais Windows n'a pas pu ouvrir automatiquement :\n"+target+"\n\nOuvre cette adresse à la main dans ton navigateur.",
			"Frieze Bender is running, but Windows couldn't open automatically:\n"+target+"\n\nOpen this address by hand in your browser."))
	}
}

// petite fenêtre d'information Windows
func alert(msg string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	m, _ := syscall.UTF16PtrFromString(msg)
	t, _ := syscall.UTF16PtrFromString(L("Courbeur de frise", "Frieze Bender"))
	proc.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x40|0x10000)
}

func openFolder(d string) { shellOpen(d) }

func uptime() time.Duration {
	k := syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")
	r, _, _ := k.Call()
	return time.Duration(r) * time.Millisecond
}

// isFrench : la langue d'affichage de Windows est-elle le français ?
func isFrench() bool {
	r, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultUILanguage").Call()
	return r&0x3ff == 0x0c
}
