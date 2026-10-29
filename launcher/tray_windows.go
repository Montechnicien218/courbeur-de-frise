//go:build windows

// Icône dans la zone de notification de Windows (à côté de l'horloge).
// Clic gauche : ouvre l'outil. Clic droit : menu.
package main

import (
	_ "embed"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

//go:embed icon.ico
var iconData []byte

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pRegisterClassEx           = user32.NewProc("RegisterClassExW")
	pCreateWindowEx            = user32.NewProc("CreateWindowExW")
	pDefWindowProc             = user32.NewProc("DefWindowProcW")
	pGetMessage                = user32.NewProc("GetMessageW")
	pTranslateMessage          = user32.NewProc("TranslateMessage")
	pDispatchMessage           = user32.NewProc("DispatchMessageW")
	pCreatePopupMenu           = user32.NewProc("CreatePopupMenu")
	pAppendMenu                = user32.NewProc("AppendMenuW")
	pTrackPopupMenu            = user32.NewProc("TrackPopupMenu")
	pDestroyMenu               = user32.NewProc("DestroyMenu")
	pGetCursorPos              = user32.NewProc("GetCursorPos")
	pSetForegroundWindow       = user32.NewProc("SetForegroundWindow")
	pPostMessage               = user32.NewProc("PostMessageW")
	pRegisterWindowMessage     = user32.NewProc("RegisterWindowMessageW")
	pLookupIconIdFromDirectory = user32.NewProc("LookupIconIdFromDirectoryEx")
	pCreateIconFromResourceEx  = user32.NewProc("CreateIconFromResourceEx")
	pGetSystemMetrics          = user32.NewProc("GetSystemMetrics")
	pLoadIcon                  = user32.NewProc("LoadIconW")
	pShellNotifyIcon           = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandle           = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmApp          = 0x8000
	wmTray         = wmApp + 1
	wmLButtonUp    = 0x0202
	wmRButtonUp    = 0x0205
	wmContextMenu  = 0x007B
	nimAdd         = 0
	nimModify      = 1
	nimDelete      = 2
	nifMessage     = 1
	nifIcon        = 2
	nifTip         = 4
	mfString       = 0
	mfSeparator    = 0x800
	tpmReturnCmd   = 0x100
	tpmNoNotify    = 0x80
	tpmRightButton = 2
)

const (
	cmdOpen = 1 + iota
	cmdForme
	cmdMotif
	cmdQuit
)

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
	_       uint32
}

var (
	trayNID        notifyIconData
	trayHwnd       uintptr
	taskbarCreated uintptr
	trayURL        string
	trayOK         bool
)

func loadTrayIcon() uintptr {
	if len(iconData) > 6 {
		cx, _, _ := pGetSystemMetrics.Call(49) // SM_CXSMICON
		cy, _, _ := pGetSystemMetrics.Call(50)
		off, _, _ := pLookupIconIdFromDirectory.Call(uintptr(unsafe.Pointer(&iconData[0])), 1, cx, cy, 0)
		if off > 0 && int(off) < len(iconData) {
			h, _, _ := pCreateIconFromResourceEx.Call(uintptr(unsafe.Pointer(&iconData[off])), uintptr(len(iconData)-int(off)), 1, 0x30000, cx, cy, 0)
			if h != 0 {
				return h
			}
		}
	}
	h, _, _ := pLoadIcon.Call(0, 32512) // IDI_APPLICATION
	return h
}

func addTrayIcon() bool {
	r, _, _ := pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&trayNID)))
	return r != 0
}

func removeTray() {
	if trayOK {
		pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&trayNID)))
		trayOK = false
	}
}

func showTrayMenu(hwnd uintptr) {
	menu, _, _ := pCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)
	add := func(id uintptr, text string) {
		t, _ := syscall.UTF16PtrFromString(text)
		pAppendMenu.Call(menu, mfString, id, uintptr(unsafe.Pointer(t)))
	}
	add(cmdOpen, L("Ouvrir le Courbeur de frise", "Open Frieze Bender"))
	add(cmdForme, L("Ouvrir le dossier Forme", "Open the Shape folder"))
	add(cmdMotif, L("Ouvrir le dossier Motif", "Open the Pattern folder"))
	pAppendMenu.Call(menu, mfSeparator, 0, 0)
	add(cmdQuit, L("Quitter", "Quit"))
	var pt struct{ X, Y int32 }
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmNoNotify|tpmRightButton, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	pPostMessage.Call(hwnd, 0, 0, 0)
	switch cmd {
	case cmdOpen:
		go openBrowser(trayURL)
	case cmdForme:
		go openFolder(inboxDirs["forme"])
	case cmdMotif:
		go openFolder(inboxDirs["motif"])
	case cmdQuit:
		exitApp()
	}
}

func trayWndProc(hwnd, umsg, wparam, lparam uintptr) uintptr {
	switch {
	case umsg == wmTray:
		switch lparam & 0xffff {
		case wmLButtonUp:
			go openBrowser(trayURL)
		case wmRButtonUp, wmContextMenu:
			showTrayMenu(hwnd)
		}
		return 0
	case taskbarCreated != 0 && umsg == taskbarCreated:
		// l'explorateur Windows a redémarré : on remet l'icône
		trayOK = addTrayIcon()
		return 0
	}
	r, _, _ := pDefWindowProc.Call(hwnd, umsg, wparam, lparam)
	return r
}

// startTray installe l'icône et fait tourner sa boucle de messages.
// ready reçoit true si l'icône est bien affichée.
func startTray(url, tip string, ready chan<- bool) {
	runtime.LockOSThread()
	trayURL = url
	inst, _, _ := pGetModuleHandle.Call(0)
	cls, _ := syscall.UTF16PtrFromString("CourbeurDeFriseTray")
	wc := wndClassEx{LpfnWndProc: syscall.NewCallback(trayWndProc), HInstance: inst, LpszClassName: cls}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	if r, _, _ := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		ready <- false
		return
	}
	title, _ := syscall.UTF16PtrFromString("Courbeur de frise")
	hwnd, _, _ := pCreateWindowEx.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), 0, 0, 0, 0, 0, 0, 0, inst, 0)
	if hwnd == 0 {
		ready <- false
		return
	}
	trayHwnd = hwnd
	tb, _ := syscall.UTF16PtrFromString("TaskbarCreated")
	taskbarCreated, _, _ = pRegisterWindowMessage.Call(uintptr(unsafe.Pointer(tb)))

	trayNID.CbSize = uint32(unsafe.Sizeof(trayNID))
	trayNID.HWnd = hwnd
	trayNID.UID = 1
	trayNID.UFlags = nifMessage | nifIcon | nifTip
	trayNID.UCallbackMessage = wmTray
	trayNID.HIcon = loadTrayIcon()
	t, _ := syscall.UTF16FromString(tip)
	copy(trayNID.SzTip[:len(trayNID.SzTip)-1], t)
	trayOK = addTrayIcon()
	ready <- trayOK

	var m msg
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func exitApp() {
	removeTray()
	logf("arrêt")
	os.Exit(0)
}
