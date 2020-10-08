package ssawsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Sync - requests sync using SYNC_URL handler passing origin parameter
// This is supposed to be called from everything that updates SH via ssawsync.Sync(origin)
func Sync(origin string) (err error) {
	// FIXME: disable for v2
	if 1 == 1 {
		return
	}
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
	url := fmt.Sprintf("%s/sync/%s", syncURL, origin)
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
	/*
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			err = e
			return
		}
		fmt.Printf("%s\n", body)
	*/
	return
}

// SyncGitdm - requests sync with gitdm (gitdmURL) with caller arg caller
// This is supposed to be called from SSAW only for selected origins
func SyncGitdm(gitdmURL, caller string) (err error) {
	// FIXME: disable for v2
	if 1 == 1 {
		return
	}
	if gitdmURL == "" || caller == "" {
		err = fmt.Errorf("gitdmURL and caller both must be set gitdmURL: %s, caller: %s", gitdmURL, caller)
		return
	}
	method := http.MethodPost
	url := fmt.Sprintf("%s/sync-from-db/%s", gitdmURL, caller)
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
	return
}
