package ssawsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Sync - requests sync using SYNC_URL handler passing origin parameter
func Sync(origin string) (err error) {
	if origin == "" {
		err = fmt.Errorf("origin cannot be empty")
		return
	}
	syncURL := os.Getenv("SYNC_URL")
	if syncURL == "" {
		err = fmt.Errorf("SYNC_URL env variable must be set")
		return
	}
	method := http.MethodPost
	url := fmt.Sprintf("http://%s/sync/%s", syncURL, origin)
	req, e := http.NewRequest(method, url, nil)
	if e != nil {
		err = fmt.Errorf("new request error: %+v for %s url: %s\n", e, method, url)
		return
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		err = fmt.Errorf("do request error: %+v for %s url: %s\n", e, method, url)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			err = fmt.Errorf("ReadAll non-ok request error: %+v for %s url: %s\n", err, method, url)
			return
		}
		err = fmt.Errorf("Method:%s url:%s status:%d\n%s\n", method, url, resp.StatusCode, body)
		return
	}
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		err = e
		return
	}
	_ = resp.Body.Close()
	fmt.Printf("%s\n", body)
	return
}
