Amp cache purging experiment - https://developers.google.com/amp/cache/update-cache

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