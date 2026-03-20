package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"flag"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

//go:embed static
var staticFiles embed.FS

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	useTLS := flag.Bool("tls", false, "enable HTTPS with a self-signed certificate")
	flag.Parse()

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fileServer := http.FileServer(http.FS(staticFS))

	testCookies := buildTestCookies()

	mux := http.NewServeMux()

	// serve static assets at their exact paths
	mux.Handle("/main.wasm", fileServer)
	mux.Handle("/wasm_exec.js", fileServer)

	// report endpoint for browser-side comparison
	mux.Handle("/api/report", handleReport(testCookies))

	// serve index.html on all other paths so the site works at any path
	indexHTML, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, c := range testCookies {
			http.SetCookie(w, c)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	scheme := "http"
	if *useTLS {
		scheme = "https"
		tlsCfg, err := selfSignedTLSConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tls: %v\n", err)
			os.Exit(1)
		}
		ln = tls.NewListener(ln, tlsCfg)
	}
	fmt.Printf("%s://%s/sub/path\n", scheme, ln.Addr())

	if err := http.Serve(ln, mux); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func buildTestCookies() []*http.Cookie {
	expires := time.Now().Add(24 * time.Hour)

	return []*http.Cookie{
		// basic
		{Name: "basic", Value: "plain", Path: "/"},

		// with expiry
		{Name: "with_expires", Value: "has_expiry", Path: "/", Expires: expires},

		// secure
		{Name: "secure_only", Value: "sec", Path: "/", Secure: true},

		// httponly (invisible to JS — document.cookie and Cookie Store API cannot see this)
		{Name: "http_only", Value: "hidden", Path: "/", HttpOnly: true},

		// samesite variations
		{Name: "ss_lax", Value: "lax_val", Path: "/", SameSite: http.SameSiteLaxMode},
		{Name: "ss_strict", Value: "strict_val", Path: "/", SameSite: http.SameSiteStrictMode},
		{Name: "ss_none", Value: "none_val", Path: "/", Secure: true, SameSite: http.SameSiteNoneMode},

		// path scoped
		{Name: "path_scoped", Value: "sub", Path: "/sub/path/"},
		{Name: "path_root", Value: "root", Path: "/"},

		// combined attributes
		{Name: "combo", Value: "all_attrs", Path: "/", Secure: true, SameSite: http.SameSiteLaxMode, Expires: expires},

		// long value
		{Name: "long_value", Value: "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", Path: "/"},

		// partitioned
		{Name: "partitioned", Value: "chip", Path: "/", Secure: true, SameSite: http.SameSiteNoneMode, Partitioned: true},

		// special characters in value
		{Name: "special_chars", Value: "hello%20world%26more%3Dstuff", Path: "/"},
	}
}

func selfSignedTLSConfig() (*tls.Config, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
