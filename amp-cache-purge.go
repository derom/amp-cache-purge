package ampcachepurge

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
	"os"
	"strings"
	"time"
)

func PurgeUrl(rawURL string) error {
	parsedUrl, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return fmt.Errorf("failed to parse url: %s", rawURL)
	}

	// make requests to cache to check it exists

	ampCDN := makeAmpCDNUrl(parsedUrl.Host)
	wpFullUrl := preparePurgingUrl(ampCDN, "wp", parsedUrl)
	cFullUrl := preparePurgingUrl(ampCDN, "c", parsedUrl)

	var err error
	// clear both wp/s and c/s
	err = makePurgeRequest(cFullUrl)
	if err != nil {
		return err
	}
	err = makePurgeRequest(wpFullUrl)
	if err != nil {
		return err
	}
	return nil
}

func makePurgeRequest(url string) error {
	fmt.Printf("Purging %s\n", url)
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

func preparePurgingUrl(ampCDN, cachePrefix string, url *url.URL) string {
	now := time.Now()
	sec := now.Unix()
	path := fmt.Sprintf("/update-cache/%s/s/%s%s?amp_action=flush&amp_ts=%d", cachePrefix, url.Host, url.RequestURI(), sec)
	signed := sign(path)
	signedEncoded := encodeSignatureForUrl(signed)

	return fmt.Sprintf("%s%s&amp_url_signature=%s", ampCDN, path, signedEncoded)
}

func makeAmpCDNUrl(host string) string {
	return fmt.Sprintf("https://%s.cdn.ampproject.org", strings.ReplaceAll(host, ".", "-"))
}

// based on https://developers.google.com/amp/cache/update-cache#rsa-keys
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
	privateKeyLocation := os.Getenv("PRIVATE_KEY_LOCATION")
	if len(privateKeyLocation) == 0 {
		privateKeyLocation = "private-key.pem"
	}
	privateKeyPassword := os.Getenv("PRIVATE_KEY_PASSWORD")
	if len(privateKeyPassword) == 0 {
		privateKeyPassword = ""
	}
	privateKey, _ := loadPrivateKey(privateKeyLocation, privateKeyPassword)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum[:])
	if err != nil {
		panic(err)
	}
	return signature
}

func loadPrivateKey(privateKeyLocation, privateKeyPassword string) (*rsa.PrivateKey, error) {
	// validate locations

	privateKeyFile, _ := ioutil.ReadFile(privateKeyLocation)

	privatePem, _ := pem.Decode(privateKeyFile)
	var privatePemBytes []byte

	if privateKeyPassword != "" {
		privatePemBytes, _ = x509.DecryptPEMBlock(privatePem, []byte(privateKeyPassword))
	} else {
		privatePemBytes = privatePem.Bytes
	}

	var parsedKey interface{}
	parsedKey, _ = x509.ParsePKCS1PrivateKey(privatePemBytes)

	var privateKey *rsa.PrivateKey
	privateKey, _ = parsedKey.(*rsa.PrivateKey)

	return privateKey, nil
}
