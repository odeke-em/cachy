package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	cachyfs "github.com/odeke-em/cachy/fs"
)

const (
	EnvCachyServerAddress = "CACHY_SERVER_ADDRESS"
	EnvCachyCertFile      = "CACHY_CERT_FILE"
	EnvCachyKeyFile       = "CACHY_KEY_FILE"
)

var fs, _ = cachyfs.New(".")

func envGetOrAlternatives(envVar string, alternatives ...string) string {
	if retr := os.Getenv(envVar); retr != "" {
		return retr
	}

	for _, alt := range alternatives {
		if alt != "" {
			return alt
		}
	}
	return ""
}

type inRequest struct {
	URI string `json:"uri"`
}

func Cache(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("%q not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uri := r.PostForm.Get("uri")
	log.Println("uri", uri, r.PostForm)
	rcs, err := fs.Open(uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := io.Copy(w, rcs)
	log.Printf("%v bytes written for uri %q err %v\n", n, uri, err)
	_ = rcs.Close()
}

func main() {
	if _, err := fs.Mount(); err != nil {
		log.Printf("fs-mounting %v\n", err)
		os.Exit(-1)
	}

	var server http.Server
	server.Addr = envGetOrAlternatives(EnvCachyServerAddress, ":8080")

	log.Printf("cachy server running on %q\n", server.Addr)

	http.HandleFunc("/", Cache)
	http.HandleFunc("/cache", Cache)

	certFilepath := envGetOrAlternatives(EnvCachyCertFile, "cachy.cert")
	keyFilepath := envGetOrAlternatives(EnvCachyKeyFile, "cachy.key")

	if err := server.ListenAndServeTLS(certFilepath, keyFilepath); err != nil {
		log.Printf("servingTLS: %v\n", err)
	}
}
