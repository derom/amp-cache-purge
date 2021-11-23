Amp cache purging experiment - https://developers.google.com/amp/cache/update-cache

## Supported Caches
Based on https://amp.dev/documentation/guides-and-tutorials/learn/amp-caches-and-cors/amp-cache-urls/
* /c - Content: This is an AMP document served as a standalone page which may be linked to directly in some interfaces.
* /v - Viewer: This is also an AMP document, but is served in an AMP Viewer which is a frame environment that displays an AMP document in the context of a Search Result Page or other interface.
* /wp - Web Package: This is an AMP document served as a Signed Exchange, a Web Package technology. These URLs act as redirects to the publisher’s own origin.

## Configuration
`PRIVATE_KEY_LOCATION` - default="private-key.pem". The private key from this repo is for tests.

`PRIVATE_KEY_PASSWORD` - default=""

## Usage example
Here is an example http server. Purging requests are slow, so it's strongly recommended to use a queue(for example, rabbitmq) instead of direct http requests.
```
package main

import (
	"net/http"

	ampcachepurge "github.com/derom/amp-cache-purge"
)

func purgeHandler(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")

	err := ampcachepurge.PurgeUrl(url)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("success"))
}

func main() {
	http.HandleFunc("/purge", purgeHandler)

	http.ListenAndServe(":8090", nil)
}
```
