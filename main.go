package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func purgeHandler(w http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(req.FormValue("url"))

	if err != nil {
		http.Error(w, "failed to parse url", http.StatusBadRequest)
		return
	}

	ampCDN := fmt.Sprintf("https://%s.cdn.ampproject.org", strings.ReplaceAll(url.Host, ".", "-"))
	now := time.Now()
	sec := now.Unix()
	wpPath := fmt.Sprintf("/update-cache/wp/s/%s%s?amp_action=flush&amp_ts=%d", url.Host, url.RequestURI(), sec)
	wpSigned := sign(wpPath)
	wpSignedEncoded := encodeSignatureForUrl(wpSigned)
	wpFullUrl := fmt.Sprintf("%s%s&amp_url_signature=%s", ampCDN, wpPath, wpSignedEncoded)

	cPath := fmt.Sprintf("/update-cache/c/s/%s%s?amp_action=flush&amp_ts=%d", url.Host, url.RequestURI(), sec)
	сSigned := sign(cPath)
	сSignedEncoded := encodeSignatureForUrl(сSigned)
	cFullUrl := fmt.Sprintf("%s%s&amp_url_signature=%s", ampCDN, cPath, сSignedEncoded)

	// clear both wp/s and c/s
	err = makePurgeRequest(w, cFullUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = makePurgeRequest(w, wpFullUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("success"))
}

func makePurgeRequest(w http.ResponseWriter, url string) error {
	fmt.Printf("Purging %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to purge %s", url)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(body) != "OK" {
		return fmt.Errorf("failed to purge %s", url)
	}
	return nil
}

func encodeSignatureForUrl(signature []byte) string {
	encoded := base64.StdEncoding.EncodeToString(signature)
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "=", "")

	return encoded
}

func sign(msg string) []byte {
	msgBytes := []byte(msg)

	msgHashSum := sha256.Sum256(msgBytes)
	privateKey, _ := loadPrivateKey("private-key.pem", "", "public-key.pem")
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum[:])
	if err != nil {
		panic(err)
	}
	return signature
}

func loadPrivateKey(rsaPrivateKeyLocation, rsaPrivateKeyPassword, rsaPublicKeyLocation string) (*rsa.PrivateKey, error) {
	// validate locations

	privateKeyFile, _ := ioutil.ReadFile(rsaPrivateKeyLocation)

	privatePem, _ := pem.Decode(privateKeyFile)
	var privatePemBytes []byte

	if rsaPrivateKeyPassword != "" {
		privatePemBytes, _ = x509.DecryptPEMBlock(privatePem, []byte(rsaPrivateKeyPassword))
	} else {
		privatePemBytes = privatePem.Bytes
	}

	var parsedKey interface{}
	parsedKey, _ = x509.ParsePKCS1PrivateKey(privatePemBytes)

	var privateKey *rsa.PrivateKey
	privateKey, _ = parsedKey.(*rsa.PrivateKey)

	return privateKey, nil
}

func main() {
	http.HandleFunc("/purge", purgeHandler)

	http.ListenAndServe(":8090", nil)
}
