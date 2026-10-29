// Courbeur de frise — lanceur Windows.
// Sert la page en local, l'ouvre dans le navigateur et relaie les envois
// vers l'API REST de LightBurn (port 19520).
package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var page []byte

const appName = "Courbeur de frise"
const appVersion = "lanceur-3"

var trayActive bool

var (
	lbBase   = envOr("LB_URL", "http://127.0.0.1:19520")
	mu       sync.Mutex
	lastPing = time.Now()
	sendMu   sync.Mutex
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

/* ---------- secret ---------- */

func secretPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "CourbeurDeFrise")
	os.MkdirAll(dir, 0o700)
	return filepath.Join(dir, "lightburn.secret")
}

func loadSecret() string {
	b, err := os.ReadFile(secretPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func saveSecret(s string) { os.WriteFile(secretPath(), []byte(s), 0o600) }
func forgetSecret()       { os.Remove(secretPath()) }

func token(secret string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(fmt.Sprint(time.Now().Unix() / 60)))
	return hex.EncodeToString(m.Sum(nil))
}

/* ---------- LightBurn ---------- */

func lbReachable() bool {
	c := http.Client{Timeout: 1500 * time.Millisecond}
	r, err := c.Get(lbBase + "/")
	if err != nil {
		return false
	}
	r.Body.Close()
	return true
}

type lbErr struct{ msg string }

func (e lbErr) Error() string { return e.msg }

func pair() (string, error) {
	body, _ := json.Marshal(map[string]any{"application_name": appName, "capabilities": []string{"state", "upload", "project"}})
	c := http.Client{Timeout: 45 * time.Second}
	r, err := c.Post(lbBase+"/api/connect", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", lbErr{"LightBurn ne répond pas. Vérifie qu'il est ouvert."}
	}
	defer r.Body.Close()
	switch r.StatusCode {
	case 200:
	case 403:
		return "", lbErr{"Autorisation refusée ou délai dépassé (30 s). Renvoie et clique « Autoriser » dans la fenêtre de LightBurn."}
	case 404:
		return "", lbErr{"Ta version de LightBurn ne propose pas encore l'API d'envoi. Mets LightBurn à jour."}
	default:
		return "", lbErr{fmt.Sprintf("LightBurn a refusé l'appairage (code %d).", r.StatusCode)}
	}
	var j struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil || j.Secret == "" {
		return "", lbErr{"Réponse d'appairage illisible."}
	}
	saveSecret(j.Secret)
	return j.Secret, nil
}

type importResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// openEvents ouvre le flux SSE et renvoie un canal qui reçoit le premier
// événement file_imported.
func openEvents(secret string) (chan importResult, func(), error) {
	req, _ := http.NewRequest("GET", lbBase+"/api/events", nil)
	req.Header.Set("Authorization", "Bearer "+token(secret))
	req.Header.Set("Accept", "text/event-stream")
	c := http.Client{}
	r, err := c.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if r.StatusCode != 200 {
		r.Body.Close()
		return nil, func() {}, lbErr{fmt.Sprint(r.StatusCode)}
	}
	ch := make(chan importResult, 1)
	go func() {
		sc := bufio.NewScanner(r.Body)
		sc.Buffer(make([]byte, 64*1024), 1<<20)
		ev := ""
		for sc.Scan() {
			l := sc.Text()
			switch {
			case strings.HasPrefix(l, "event:"):
				ev = strings.TrimSpace(l[6:])
			case strings.HasPrefix(l, "data:") && ev == "file_imported":
				var res importResult
				json.Unmarshal([]byte(strings.TrimSpace(l[5:])), &res)
				select {
				case ch <- res:
				default:
				}
				return
			case l == "":
				ev = ""
			}
		}
	}()
	return ch, func() { r.Body.Close() }, nil
}

type placement struct {
	set  bool
	x, y string
}

func upload(secret, name string, data []byte, group bool, pl placement) (int, error) {
	req, _ := http.NewRequest("POST", lbBase+"/api/file/upload", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token(secret))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Filename", name)
	if group {
		req.Header.Set("X-Group-Shapes", "true")
	}
	if pl.set {
		req.Header.Set("X-Position-X", pl.x)
		req.Header.Set("X-Position-Y", pl.y)
		req.Header.Set("X-Origin", "bottom-left")
	}
	c := http.Client{Timeout: 30 * time.Second}
	r, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	return r.StatusCode, nil
}

func send(name string, data []byte, group bool, pl placement) (string, error) {
	sendMu.Lock()
	defer sendMu.Unlock()
	if !lbReachable() {
		return "", lbErr{"LightBurn n'est pas ouvert (ou son API n'est pas disponible). Ouvre LightBurn puis réessaie."}
	}
	for attempt := 0; attempt < 2; attempt++ {
		secret := loadSecret()
		if secret == "" {
			var err error
			if secret, err = pair(); err != nil {
				return "", err
			}
		}
		events, closeEv, evErr := openEvents(secret)
		code, err := upload(secret, name, data, group, pl)
		if err != nil {
			closeEv()
			return "", lbErr{"L'envoi vers LightBurn a échoué : " + err.Error()}
		}
		if code == 401 {
			closeEv()
			forgetSecret()
			continue // ré-appairage
		}
		if code != 202 && code != 200 {
			closeEv()
			return "", lbErr{fmt.Sprintf("LightBurn a refusé le fichier (code %d).", code)}
		}
		if evErr != nil {
			return "Fichier envoyé à LightBurn.", nil
		}
		defer closeEv()
		select {
		case res := <-events:
			if res.Success {
				return "C'est dans LightBurn : la frise a été importée dans le projet ouvert.", nil
			}
			msg := "LightBurn n'a pas pu importer le fichier."
			if res.Error != "" {
				msg += " " + res.Error
			}
			return "", lbErr{msg}
		case <-time.After(20 * time.Second):
			return "Fichier envoyé. LightBurn n'a pas encore confirmé l'import : vérifie le projet.", nil
		}
	}
	return "", lbErr{"LightBurn refuse l'autorisation. Clique « Oublier l'appairage » puis réessaie."}
}

/* ---------- serveur local ---------- */

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func ping() { mu.Lock(); lastPing = time.Now(); mu.Unlock() }

func main() {
	addr := envOr("LISTEN", "127.0.0.1:19580")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// une autre copie tourne déjà : si c'est la même version, on ouvre sa page
		if runningVersion(addr) == appVersion {
			openBrowser("http://" + addr + "/")
			time.Sleep(time.Second)
			return
		}
		// autre version : on lui demande poliment de s'arrêter
		askQuit(addr)
		for i := 0; i < 20 && err != nil; i++ {
			time.Sleep(250 * time.Millisecond)
			ln, err = net.Listen("tcp", addr)
		}
		if err != nil {
			v := runningVersion(addr)
			logf("port %s occupé (version en place : %q) : %v", addr, v, err)
			if v != "" {
				alert(L("Une ancienne version du Courbeur de frise (v"+v+") tourne encore et occupe l'adresse 127.0.0.1:19580.\n\n"+
					"Arrête « CourbeurDeFrise.exe » dans le Gestionnaire des tâches (Ctrl+Maj+Échap), ou redémarre l'ordinateur, puis relance cette version.",
					"An older version of Frieze Bender (v"+v+") is still running and uses the address 127.0.0.1:19580.\n\n"+
						"End « CourbeurDeFrise.exe » in Task Manager (Ctrl+Shift+Esc), or restart the computer, then start this version again."))
			} else {
				alert(L("Impossible de démarrer : l'adresse 127.0.0.1:19580 est déjà utilisée par un autre programme.\n\nRedémarre l'ordinateur puis relance le Courbeur de frise.",
					"Can't start: the address 127.0.0.1:19580 is already used by another program.\n\nRestart the computer, then start Frieze Bender again."))
			}
			return
		}
	}
	logf("démarré, version %s, adresse %s", appVersion, addr)
	initInbox()
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if len(iconData) == 0 {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(iconData)
	})
	mux.HandleFunc("/lb/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": appVersion})
	})
	mux.HandleFunc("/lb/quit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Origin") != "" {
			http.Error(w, "refusé", 403)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
		go func() { time.Sleep(200 * time.Millisecond); exitApp() }()
	})
	mux.HandleFunc("/lightburn", func(w http.ResponseWriter, r *http.Request) { handoff(w, r, addr) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		logf("page demandée, navigateur : %s", r.UserAgent())
		// ouvert depuis le navigateur intégré de LightBurn (Qt WebEngine) :
		// on bascule vers le navigateur de l'ordinateur
		if strings.Contains(r.UserAgent(), "QtWebEngine") && r.URL.Query().Get("ici") == "" {
			handoff(w, r, addr)
			return
		}
		ping()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(currentPage())
	})
	mux.HandleFunc("/lb/status", func(w http.ResponseWriter, r *http.Request) {
		ping()
		writeJSON(w, map[string]bool{"lightburn": lbReachable(), "paired": loadSecret() != ""})
	})
	mux.HandleFunc("/lb/forget", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST", 405)
			return
		}
		forgetSecret()
		writeJSON(w, map[string]bool{"ok": true})
	})
	mux.HandleFunc("/lb/send", func(w http.ResponseWriter, r *http.Request) {
		ping()
		if r.Method != "POST" {
			http.Error(w, "POST", 405)
			return
		}
		// refuse les requêtes venant d'autres sites
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
			http.Error(w, "origine refusée", 403)
			return
		}
		data, err := io.ReadAll(io.LimitReader(r.Body, 200<<20))
		if err != nil || len(data) == 0 {
			writeJSON(w, map[string]any{"ok": false, "message": "Fichier vide."})
			return
		}
		name := filepath.Base(r.URL.Query().Get("name"))
		if name == "" || name == "." || name == "/" {
			name = "frise.dxf"
		}
		q := r.URL.Query()
		var pl placement
		if x, y := q.Get("x"), q.Get("y"); isNum(x) && isNum(y) {
			pl = placement{true, x, y}
		}
		msg, err := send(name, data, q.Get("group") == "1", pl)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "message": msg})
	})

	mux.HandleFunc("/lb/inbox", func(w http.ResponseWriter, r *http.Request) {
		ping()
		inboxMu.Lock()
		defer inboxMu.Unlock()
		writeJSON(w, map[string]any{"dirs": inboxDirs, "files": inboxLatest})
	})
	mux.HandleFunc("/lb/inbox/file", func(w http.ResponseWriter, r *http.Request) {
		kind := r.URL.Query().Get("kind")
		inboxMu.Lock()
		f, ok := inboxLatest[kind]
		inboxMu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		b, err := os.ReadFile(f.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(b)
	})
	mux.HandleFunc("/lb/opendir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST", 405)
			return
		}
		if d, ok := inboxDirs[r.URL.Query().Get("kind")]; ok {
			openFolder(d)
		}
		writeJSON(w, map[string]bool{"ok": true})
	})
	mux.HandleFunc("/lb/layers", func(w http.ResponseWriter, r *http.Request) {
		secret := loadSecret()
		if secret == "" {
			writeJSON(w, map[string]any{"ok": false, "reason": "unpaired"})
			return
		}
		req, _ := http.NewRequest("GET", lbBase+"/api/layers", nil)
		req.Header.Set("Authorization", "Bearer "+token(secret))
		c := http.Client{Timeout: 3 * time.Second}
		resp, err := c.Do(req)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "unreachable"})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			writeJSON(w, map[string]any{"ok": false, "reason": fmt.Sprint(resp.StatusCode)})
			return
		}
		var j struct {
			Layers json.RawMessage `json:"layers"`
		}
		if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&j) != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "json"})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "layers": j.Layers})
	})
	mux.HandleFunc("/lb/prefs", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(appDir(), "reglages.json")
		if r.Method == "POST" {
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				http.Error(w, "origine refusée", 403)
				return
			}
			b, _ := io.ReadAll(io.LimitReader(r.Body, 64<<10))
			if json.Valid(b) {
				os.WriteFile(p, b, 0o600)
			}
			writeJSON(w, map[string]bool{"ok": true})
			return
		}
		b, err := os.ReadFile(p)
		if err != nil || !json.Valid(b) {
			b = []byte("{}")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(b)
	})
	mux.HandleFunc("/lb/library", func(w http.ResponseWriter, r *http.Request) {
		type item struct {
			Name  string `json:"name"`
			MTime int64  `json:"mtime"`
		}
		list := []item{}
		ents, _ := os.ReadDir(inboxDirs["motif"])
		for _, e := range ents {
			if e.IsDir() || strings.ToLower(filepath.Ext(e.Name())) != ".svg" {
				continue
			}
			if info, err := e.Info(); err == nil {
				list = append(list, item{e.Name(), info.ModTime().UnixMilli()})
			}
		}
		sort.Slice(list, func(i, j int) bool { return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name) })
		writeJSON(w, map[string]any{"files": list})
	})
	mux.HandleFunc("/lb/library/file", func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(r.URL.Query().Get("name"))
		if strings.ToLower(filepath.Ext(name)) != ".svg" {
			http.NotFound(w, r)
			return
		}
		b, err := os.ReadFile(filepath.Join(inboxDirs["motif"], name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(b)
	})
	mux.HandleFunc("/lb/library/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST", 405)
			return
		}
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
			http.Error(w, "origine refusée", 403)
			return
		}
		b, _ := io.ReadAll(io.LimitReader(r.Body, 10<<20))
		if len(b) == 0 || !bytes.Contains(bytes.ToLower(b[:min(len(b), 4096)]), []byte("<svg")) {
			writeJSON(w, map[string]any{"ok": false, "message": "Ce motif n'est pas un SVG."})
			return
		}
		base := strings.TrimSuffix(filepath.Base(r.URL.Query().Get("name")), filepath.Ext(r.URL.Query().Get("name")))
		base = strings.Map(func(c rune) rune {
			if strings.ContainsRune(`<>:"/\|?*`, c) || c < 32 {
				return '_'
			}
			return c
		}, base)
		if base == "" || base == "." {
			base = "motif"
		}
		dir := inboxDirs["motif"]
		name := base + ".svg"
		if existing, err := os.ReadFile(filepath.Join(dir, name)); err == nil && !bytes.Equal(existing, b) {
			for i := 2; ; i++ {
				name = fmt.Sprintf("%s (%d).svg", base, i)
				if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
					break
				}
			}
		}
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			writeJSON(w, map[string]any{"ok": false, "message": "Écriture impossible : " + err.Error()})
			return
		}
		info, _ := os.Stat(p)
		scanInbox()
		writeJSON(w, map[string]any{"ok": true, "name": name, "mtime": info.ModTime().UnixMilli()})
	})
	go watchInbox()

	// reste actif tant que LightBurn tourne ou que la page est ouverte
	go func() {
		for range time.Tick(30 * time.Second) {
			mu.Lock()
			idle := time.Since(lastPing)
			mu.Unlock()
			if !trayActive && idle > 10*time.Minute && !lbReachable() {
				exitApp()
			}
		}
	}()

	url := "http://" + addr + "/"
	ready := make(chan bool, 1)
	go startTray(url, L("Courbeur de frise — clic : ouvrir, clic droit : menu", "Frieze Bender — click: open, right-click: menu"), ready)
	select {
	case trayActive = <-ready:
	case <-time.After(3 * time.Second):
	}
	logf("icône dans la barre des tâches : %v", trayActive)
	if discreetStart() {
		logf("démarrage discret : navigateur non ouvert")
	} else {
		go openBrowser(url)
	}
	http.Serve(ln, mux)
}

func openBrowser(url string) {
	time.Sleep(150 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		shellOpen(url)
		return
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

/* ---------- dossier d'échange ---------- */

type inboxFile struct {
	Name  string `json:"name"`
	MTime int64  `json:"mtime"` // ms
	Path  string `json:"-"`
}

var (
	inboxMu     sync.Mutex
	inboxDirs   = map[string]string{}
	inboxLatest = map[string]inboxFile{}
)

func initInbox() {
	home, _ := os.UserHomeDir()
	docs := filepath.Join(home, "Documents")
	if st, err := os.Stat(docs); err != nil || !st.IsDir() {
		docs = home
	}
	frBase := filepath.Join(docs, "Courbeur de frise")
	names := map[string]string{"forme": "Forme", "motif": "Motif"}
	base := frBase
	if _, err := os.Stat(frBase); err != nil && !isFrench() {
		base = filepath.Join(docs, "Frieze Bender")
		names = map[string]string{"forme": "Shape", "motif": "Pattern"}
	}
	base = envOr("INBOX", base)
	for k, n := range names {
		d := filepath.Join(base, n)
		os.MkdirAll(d, 0o755)
		inboxDirs[k] = d
	}
	scanInbox()
}

func scanInbox() {
	for kind, dir := range inboxDirs {
		ents, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		var best inboxFile
		for _, e := range ents {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if e.IsDir() || (ext != ".dxf" && ext != ".svg") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			mt := info.ModTime().UnixMilli()
			if mt > best.MTime {
				best = inboxFile{Name: e.Name(), MTime: mt, Path: filepath.Join(dir, e.Name())}
			}
		}
		inboxMu.Lock()
		if best.Path != "" {
			inboxLatest[kind] = best
		} else {
			delete(inboxLatest, kind)
		}
		inboxMu.Unlock()
	}
}

func watchInbox() {
	for range time.Tick(time.Second) {
		scanInbox()
	}
}

/* ---------- bascule depuis LightBurn ---------- */

var (
	handoffMu   sync.Mutex
	lastHandoff time.Time
)

const handoffPageFR = `<!doctype html><html lang="fr"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>Courbeur de frise</title>
<style>
:root{--bg:#e9ede9;--fg:#18211d;--muted:#57665e;--accent:#16795a}
@media (prefers-color-scheme:dark){:root{--bg:#101513;--fg:#e2eae5;--muted:#93a39a;--accent:#3fbf8f}}
body{margin:0;min-height:100vh;display:grid;place-items:center;background:var(--bg);color:var(--fg);font:15px/1.5 "Segoe UI",system-ui,sans-serif}
main{max-width:420px;padding:24px;text-align:center;display:grid;gap:12px}
h1{font-size:1.25rem;margin:0}p{margin:0;color:var(--muted)}
button{font:inherit;font-weight:600;padding:8px 18px;border-radius:8px;border:0;background:var(--accent);color:#fff;cursor:pointer}
a{color:var(--muted);font-size:.85rem}
</style></head><body><main>
<h1>Courbeur de frise ouvert dans ton navigateur</h1>
<p>Cette fenêtre va se fermer. Si elle reste ouverte, ferme-la : tu retrouves l'outil dans ton navigateur habituel.</p>
<button onclick="window.close()">Fermer cette fenêtre</button>
<p>Rien ne s'est ouvert ? <a href="http://localhost:19580/" target="_blank" rel="noopener">Ouvrir dans mon navigateur</a></p>
<a href="/?ici=1">Utiliser quand même l'outil ici</a>
<small style="color:var(--muted)">bascule</small>
</main><script>setTimeout(function(){try{window.close()}catch(e){}},900)</script></body></html>`

const handoffPageEN = `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>Frieze Bender</title>
<style>
:root{--bg:#e9ede9;--fg:#18211d;--muted:#57665e;--accent:#16795a}
@media (prefers-color-scheme:dark){:root{--bg:#101513;--fg:#e2eae5;--muted:#93a39a;--accent:#3fbf8f}}
body{margin:0;min-height:100vh;display:grid;place-items:center;background:var(--bg);color:var(--fg);font:15px/1.5 "Segoe UI",system-ui,sans-serif}
main{max-width:420px;padding:24px;text-align:center;display:grid;gap:12px}
h1{font-size:1.25rem;margin:0}p{margin:0;color:var(--muted)}
button{font:inherit;font-weight:600;padding:8px 18px;border-radius:8px;border:0;background:var(--accent);color:#fff;cursor:pointer}
a{color:var(--muted);font-size:.85rem}
</style></head><body><main>
<h1>Frieze Bender opened in your browser</h1>
<p>This window will close. If it stays open, close it: the tool is in your usual browser.</p>
<button onclick="window.close()">Close this window</button>
<p>Nothing opened? <a href="http://localhost:19580/" target="_blank" rel="noopener">Open in my browser</a></p>
<a href="/?ici=1">Use the tool here anyway</a>
<small style="color:var(--muted)">handoff</small>
</main><script>setTimeout(function(){try{window.close()}catch(e){}},900)</script></body></html>`

func handoff(w http.ResponseWriter, r *http.Request, addr string) {
	ping()
	logf("bascule demandée par LightBurn")
	handoffMu.Lock()
	if time.Since(lastHandoff) > 3*time.Second { // évite les doubles ouvertures
		lastHandoff = time.Now()
		go openBrowser("http://" + addr + "/")
	}
	handoffMu.Unlock()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	pg := handoffPageEN
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Accept-Language"))), "fr") || (r.Header.Get("Accept-Language") == "" && isFrench()) {
		pg = handoffPageFR
	}
	w.Write([]byte(pg))
}

func runningVersion(addr string) string {
	c := http.Client{Timeout: 1500 * time.Millisecond}
	r, err := c.Get("http://" + addr + "/lb/version")
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	var j struct {
		Version string `json:"version"`
	}
	json.NewDecoder(r.Body).Decode(&j)
	return j.Version
}

func askQuit(addr string) {
	c := http.Client{Timeout: 1500 * time.Millisecond}
	r, err := c.Post("http://"+addr+"/lb/quit", "text/plain", nil)
	if err == nil {
		r.Body.Close()
	}
}

/* ---------- journal ---------- */

func logf(format string, a ...any) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, "CourbeurDeFrise")
	os.MkdirAll(dir, 0o700)
	f, err := os.OpenFile(filepath.Join(dir, "journal.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s  %s\r\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, a...))
}

func isNum(v string) bool {
	if v == "" || len(v) > 20 {
		return false
	}
	var f float64
	_, err := fmt.Sscan(v, &f)
	return err == nil
}

func appDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "CourbeurDeFrise")
	os.MkdirAll(dir, 0o700)
	return dir
}

// pas de navigateur au lancement si on le demande (--discret) ou juste après le démarrage de Windows
func discreetStart() bool {
	for _, a := range os.Args[1:] {
		a = strings.ToLower(strings.TrimLeft(a, "-/"))
		if a == "discret" || a == "silencieux" {
			return true
		}
	}
	return uptime() > 0 && uptime() < 4*time.Minute
}

// La page est lue à côté du programme (courbeur.html) : pour passer à une nouvelle version
// de l'outil, il suffit de remplacer ce fichier, sans changer de .exe.
// Sans ce fichier, la page intégrée au programme est utilisée.
func currentPage() []byte {
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "courbeur.html")
		if b, err := os.ReadFile(p); err == nil && len(b) > 1000 && bytes.Contains(b[:min(len(b), 4096)], []byte("Courbeur de frise")) {
			return b
		}
	}
	return page
}

// L choisit le texte selon la langue de Windows (français ou anglais).
func L(fr, en string) string {
	if isFrench() {
		return fr
	}
	return en
}
