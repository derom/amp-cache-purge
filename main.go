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

func purge(w http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(req.FormValue("url"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to parse url"))
	}

	ampCDN := fmt.Sprintf("https://%s.cdn.ampproject.org", strings.ReplaceAll(url.Host, ".", "-"))
	now := time.Now()
	sec := now.Unix()
	wpPath := fmt.Sprintf("/update-cache/wp/s/%s%s?amp_action=flush&amp_ts=%d", url.Host, url.RequestURI(), sec)
	wpSigned := sign(wpPath)
	wpSignedEncoded := encodeSignatureForUrl(wpSigned)
	wpFull := fmt.Sprintf("%s&amp_url_signature=%s", wpPath, wpSignedEncoded)

	cPath := fmt.Sprintf("/update-cache/c/s/%s%s?amp_action=flush&amp_ts=%d", url.Host, url.RequestURI(), sec)
	сSigned := sign(cPath)
	сSignedEncoded := encodeSignatureForUrl(сSigned)
	cFull := fmt.Sprintf("%s&amp_url_signature=%s", cPath, сSignedEncoded)

	// clear both wp/s and c/s
	fmt.Fprintf(w, "%s%s\n", ampCDN, wpFull)
	fmt.Fprintf(w, "%s%s\n", ampCDN, cFull)
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
	privateKey, _ := rsaConfigSetup("private-key.pem", "", "public-key.pem")
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum[:])
	if err != nil {
		panic(err)
	}
	return signature
}

func rsaConfigSetup(rsaPrivateKeyLocation, rsaPrivateKeyPassword, rsaPublicKeyLocation string) (*rsa.PrivateKey, error) {
	// validate locations

	priv, _ := ioutil.ReadFile(rsaPrivateKeyLocation)

	privPem, _ := pem.Decode(priv)
	var privPemBytes []byte

	if rsaPrivateKeyPassword != "" {
		privPemBytes, _ = x509.DecryptPEMBlock(privPem, []byte(rsaPrivateKeyPassword))
	} else {
		privPemBytes = privPem.Bytes
	}

	var parsedKey interface{}
	var err error
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			//utils.LogError("Unable to parse RSA private key, generating a temp one", err, utils.LogFields{})
			//return GenRSA(4096)
		}
	}

	var privateKey *rsa.PrivateKey
	privateKey, _ = parsedKey.(*rsa.PrivateKey)

	pub, _ := ioutil.ReadFile(rsaPublicKeyLocation)

	pubPem, _ := pem.Decode(pub)

	parsedKey, _ = x509.ParsePKIXPublicKey(pubPem.Bytes)

	var pubKey *rsa.PublicKey
	pubKey, _ = parsedKey.(*rsa.PublicKey)

	privateKey.PublicKey = *pubKey

	return privateKey, nil
}

func main() {

	http.HandleFunc("/purge", purge)

	http.ListenAndServe(":8090", nil)
}
