package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	serverName   = "localhost"
	SSLport      = ":1443"
	HTTPport     = ":8080"
	SSLprotocol  = "https://"
	HTTPprotocol = "http://"
)

func secureRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "You have arrived at port 443, and now you are marginally more secure.")
}

func redirectNonSecure(w http.ResponseWriter, r *http.Request) {
	log.Println("Non-secure request initiated, redirecting.")
	redirectURL := SSLprotocol + serverName + SSLport + r.RequestURI
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

const randomLength = 45

func GenerateSalt(length int) string {
	var salt []byte
	var asciiPad int64

	if length == 0 {
		length = randomLength
	}

	asciiPad = 32

	for i := 0; i < length; i++ {
		salt = append(salt, byte(rand.Int63n(94)+asciiPad))
	}
	return string(salt)
}

func GenerateHash(salt string, password string) string {
	var hash string
	fullString := salt + password
	sha := sha256.New()
	sha.Write([]byte(fullString))
	hash = base64.URLEncoding.EncodeToString(sha.Sum(nil))
	return hash
}

func ReturnPassword(password string) (string, string) {
	rand.Seed(time.Now().UTC().UnixNano())
	salt := GenerateSalt(0)
	hash := GenerateHash(salt, password)
	return salt, hash
}

func main() {
	salt, hash := ReturnPassword("password")
	fmt.Println("password salt and hash:", salt, hash)
	wg := sync.WaitGroup{}
	log.Println("Starting redirection server, try to access @ http:")

	wg.Add(1)
	go func() {
		http.ListenAndServe(HTTPport, http.HandlerFunc(redirectNonSecure))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		http.ListenAndServeTLS(SSLport, "server.pem", "server.key", http.HandlerFunc(secureRequest))
		wg.Done()
	}()
	wg.Wait()
}
