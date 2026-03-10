package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

type fileState struct {
	mu      sync.RWMutex
	content string
	changed bool
	lastMod time.Time
}

func (f *fileState) update(content string, modTime time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.content = content
	f.changed = true
	f.lastMod = modTime
}

func (f *fileState) read() (string, bool, time.Time) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.content, f.changed, f.lastMod
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %q is not set", key)
	}
	return v
}

// hashCredential hashes a credential with SHA-256 so ConstantTimeCompare
// never operates on the raw secret string.
func hashCredential(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

func corsMiddleware(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin == "*" || origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func basicAuth(expectedUser, expectedPass string) func(http.HandlerFunc) http.HandlerFunc {
	userHash := hashCredential(expectedUser)
	passHash := hashCredential(expectedPass)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="apiwatch"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userMatch := subtle.ConstantTimeCompare(hashCredential(user), userHash)
			passMatch := subtle.ConstantTimeCompare(hashCredential(pass), passHash)

			// Both comparisons must run before branching to avoid timing leaks
			if (userMatch & passMatch) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="apiwatch"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}

func watchFile(path string, state *fileState) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Watch the directory to catch atomic writes (editors often replace the file)
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		log.Fatalf("failed to watch directory: %v", err)
	}

	log.Printf("watching file: %s", path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Clean(event.Name) != filepath.Clean(path) {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				data, err := os.ReadFile(path)
				if err != nil {
					log.Printf("error reading file: %v", err)
					continue
				}
				info, err := os.Stat(path)
				if err != nil {
					log.Printf("error stating file: %v", err)
					continue
				}
				state.update(string(data), info.ModTime())
				log.Printf("file changed: %s at %s", path, info.ModTime().Format(time.RFC3339))
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, falling back to shell environment")
	}

	authUser := mustEnv("APIWATCH_USER")
	authPass := mustEnv("APIWATCH_PASS")
	serverPort := mustEnv("APIWATCH_PORT")
	allowedOrigin := os.Getenv("APIWATCH_CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}
	auth := basicAuth(authUser, authPass)

	_, filename, _, _ := runtime.Caller(0)
	moduleDir := filepath.Dir(filename)
	collectionPath := filepath.Join(moduleDir, "..", "resource", "apidocs", "collection.json")
	collectionPath = filepath.Clean(collectionPath)

	state := &fileState{}

	// Load initial content
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		log.Fatalf("failed to read initial file: %v", err)
	}
	info, err := os.Stat(collectionPath)
	if err != nil {
		log.Fatalf("failed to stat initial file: %v", err)
	}
	state.update(string(data), info.ModTime())
	state.mu.Lock()
	state.changed = false // not changed on initial load
	state.mu.Unlock()

	go watchFile(collectionPath, state)

	mux := http.NewServeMux()
	mux.HandleFunc("/collection", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		content, changed, lastMod := state.read()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"changed":  changed,
			"last_mod": lastMod.Format(time.RFC3339),
			"content":  json.RawMessage(content),
		})
	}))

	log.Printf("apiwatch server running on %s (CORS origin: %s)", serverPort, allowedOrigin)
	log.Fatal(http.ListenAndServe(serverPort, corsMiddleware(allowedOrigin, mux)))
}
