package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

func parseInRequest(req *http.Request) (*inRequest, error) {
	var body io.Reader
	switch req.Method {
	case "GET":
		values := req.URL.Query()
		kvList := []string{}
		for key, _ := range values {
			// TODO: Fix this hack since not only string values will be passed in
			kvList = append(kvList, fmt.Sprintf("%q:%q", key, values.Get(key)))
			fakeJSON := fmt.Sprintf("{%s}", strings.Join(kvList, ","))
			body = strings.NewReader(fakeJSON)
		}
	case "POST", "PUT": // Lax method acceptance
		limBody := io.LimitReader(req.Body, 1<<20)
		body = limBody
	default:
		return nil, fmt.Errorf("Method %q not supported", req.Method)
	}

	if body == nil {
		return nil, fmt.Errorf("bug on: body is nil")
	}

	slurp, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	inReq := &inRequest{}
	if err := json.Unmarshal(slurp, inReq); err != nil {
		return nil, err
	}
	return inReq, nil
}

func Cache(w http.ResponseWriter, r *http.Request) {
	inReq, err := parseInRequest(r)
	// Close the body first
	_ = r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	uri := strings.TrimSpace(inReq.URI)
	if uri == "" {
		http.Error(w, "'uri' expected in postForm", http.StatusBadRequest)
		return
	}

	log.Println("uri", uri, r.PostForm)
	rcs, err := fs.Open(uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := io.Copy(w, rcs)
	log.Printf("%v bytes sent out for uri %q err %v\n", n, uri, err)
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

	if envGetOrAlternatives("http1") != "" {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("serving http: %v\n", err)
		}
	} else {
		if err := server.ListenAndServeTLS(certFilepath, keyFilepath); err != nil {
			log.Printf("servingTLS: %v\n", err)
		}
	}
}
