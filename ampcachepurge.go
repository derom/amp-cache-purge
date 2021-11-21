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

	ampCDN := makeAmpCDNUrl(parsedUrl.Host)
	now := time.Now()
	timestamp := now.Unix()

	// https://amp.dev/documentation/guides-and-tutorials/learn/amp-caches-and-cors/amp-cache-urls/
	// webpackages (SXG)
	wpFullUrl := preparePurgingUrl(ampCDN, "wp", parsedUrl, timestamp)
	// content
	cFullUrl := preparePurgingUrl(ampCDN, "c", parsedUrl, timestamp)
	// viewer
	// check viewer cache exists
	vCacheUrl := prepareCacheUrl(ampCDN, "v", parsedUrl)
	cacheExists := checkCacheExists(vCacheUrl)
	var err error
	if cacheExists {
		vFullUrl := preparePurgingUrl(ampCDN, "v", parsedUrl, timestamp)
		err = makePurgeRequest(vFullUrl)
		if err != nil {
			return err
		}
	}

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

func checkCacheExists(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}

func makePurgeRequest(url string) error {
	fmt.Printf("Purging %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("failed to purge %s", url)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(body) != "OK" {
		fmt.Println(err.Error())
		return fmt.Errorf("failed to purge %s", url)
	}
	return nil
}

func preparePurgingUrl(ampCDN, cachePrefix string, url *url.URL, timestamp int64) string {
	path := fmt.Sprintf("/update-cache/%s/s/%s%s?amp_action=flush&amp_ts=%d", cachePrefix, url.Host, url.RequestURI(), timestamp)
	signed := sign(path)
	signedEncoded := encodeSignatureForUrl(signed)

	return fmt.Sprintf("%s%s&amp_url_signature=%s", ampCDN, path, signedEncoded)
}

func prepareCacheUrl(ampCDN, cachePrefix string, url *url.URL) string {
	return fmt.Sprintf("%s/%s/s/%s%s", ampCDN, cachePrefix, url.Host, url.RequestURI())
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
	privateKey, err := loadPrivateKey(privateKeyLocation, privateKeyPassword)
	if err != nil {
		panic(err)
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum[:])
	if err != nil {
		panic(err)
	}
	return signature
}

func loadPrivateKey(privateKeyLocation, privateKeyPassword string) (*rsa.PrivateKey, error) {
	// validate locations

	privateKeyFile, err := ioutil.ReadFile(privateKeyLocation)
	if err != nil {
		return nil, err
	}

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
